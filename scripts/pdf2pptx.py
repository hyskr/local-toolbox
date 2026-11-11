#!/usr/bin/env python3
"""Convert PDF pages to picture slides and preserve Touying speaker notes."""

import argparse
import json
import os
import re
import struct
import subprocess
import tempfile
import zipfile
from pathlib import Path
from xml.sax.saxutils import escape


P = "http://schemas.openxmlformats.org/presentationml/2006/main"
A = "http://schemas.openxmlformats.org/drawingml/2006/main"
R = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
REL = "http://schemas.openxmlformats.org/package/2006/relationships"
CT = "http://schemas.openxmlformats.org/package/2006/content-types"
XML = '<?xml version="1.0" encoding="UTF-8" standalone="yes"?>'
NS = f'xmlns:p="{P}" xmlns:a="{A}" xmlns:r="{R}"'
GROUP = '<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="0" cy="0"/><a:chOff x="0" y="0"/><a:chExt cx="0" cy="0"/></a:xfrm></p:grpSpPr>'
COLORS = 'bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"'


def relationships(items):
    return XML + f'<Relationships xmlns="{REL}">' + ''.join(
        f'<Relationship Id="rId{i}" Type="{R}/{kind}" Target="{target}"/>'
        for i, kind, target in items
    ) + '</Relationships>'


def theme():
    colors = {"dk1": "202034", "lt1": "FFFFFF", "dk2": "2F2F7F", "lt2": "E8E2F4",
              "accent1": "2F2F7F", "accent2": "805DE0", "accent3": "56566E",
              "accent4": "4D4DA3", "accent5": "B7A5DF", "accent6": "D4D4DE",
              "hlink": "0563C1", "folHlink": "954F72"}
    palette = ''.join(f'<a:{k}><a:srgbClr val="{v}"/></a:{k}>' for k, v in colors.items())
    fonts = ''.join(f'<a:{kind}Font><a:latin typeface="Arial"/><a:ea typeface="Source Han Sans CN"/><a:cs typeface="Arial"/></a:{kind}Font>' for kind in ("major", "minor"))
    fills = '<a:solidFill><a:schemeClr val="phClr"/></a:solidFill>' * 3
    lines = '<a:ln w="9525"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln>' * 3
    return XML + f'<a:theme xmlns:a="{A}" name="Slides"><a:themeElements><a:clrScheme name="Academic">{palette}</a:clrScheme><a:fontScheme name="Slides">{fonts}</a:fontScheme><a:fmtScheme name="Slides"><a:fillStyleLst>{fills}</a:fillStyleLst><a:lnStyleLst>{lines}</a:lnStyleLst><a:effectStyleLst>{"<a:effectStyle><a:effectLst/></a:effectStyle>" * 3}</a:effectStyleLst><a:bgFillStyleLst>{fills}</a:bgFillStyleLst></a:fmtScheme></a:themeElements></a:theme>'


def notes_xml(text):
    paragraphs = ''.join('<a:p><a:r><a:rPr lang="zh-CN"/><a:t xml:space="preserve">'
                         + escape(line) + '</a:t></a:r></a:p>' for line in text.split('\n'))
    return XML + f'<p:notes {NS}><p:cSld><p:spTree>{GROUP}<p:sp><p:nvSpPr><p:cNvPr id="2" name="Notes"/><p:cNvSpPr/><p:nvPr><p:ph type="body" idx="1"/></p:nvPr></p:nvSpPr><p:spPr/><p:txBody><a:bodyPr/><a:lstStyle/>{paragraphs}</p:txBody></p:sp></p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:notes>'


def picture_xml(width, height, image_width, image_height):
    scale = min(width / image_width, height / image_height)
    cx, cy = round(image_width * scale), round(image_height * scale)
    x, y = (width - cx) // 2, (height - cy) // 2
    picture = f'<p:pic><p:nvPicPr><p:cNvPr id="2" name="PDF page"/><p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr><p:nvPr/></p:nvPicPr><p:blipFill><a:blip r:embed="rId2"/><a:stretch><a:fillRect/></a:stretch></p:blipFill><p:spPr><a:xfrm><a:off x="{x}" y="{y}"/><a:ext cx="{cx}" cy="{cy}"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr></p:pic>'
    return XML + f'<p:sld {NS}><p:cSld><p:spTree>{GROUP}{picture}</p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sld>'


def write_pptx(output, images, notes, width, height):
    overrides = {"/ppt/presentation.xml": "presentation.main",
                 "/ppt/slideMasters/slideMaster1.xml": "slideMaster",
                 "/ppt/slideLayouts/slideLayout1.xml": "slideLayout",
                 "/ppt/notesMasters/notesMaster1.xml": "notesMaster"}
    for i in range(1, len(images) + 1):
        overrides[f'/ppt/slides/slide{i}.xml'] = 'slide'
        overrides[f'/ppt/notesSlides/notesSlide{i}.xml'] = 'notesSlide'
    types = XML + f'<Types xmlns="{CT}"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Default Extension="png" ContentType="image/png"/><Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>'
    types += ''.join(f'<Override PartName="{name}" ContentType="application/vnd.openxmlformats-officedocument.presentationml.{kind}+xml"/>' for name, kind in overrides.items()) + '</Types>'
    slide_ids = ''.join(f'<p:sldId id="{256+i}" r:id="rId{i+3}"/>' for i in range(len(images)))
    with zipfile.ZipFile(output, 'w', zipfile.ZIP_DEFLATED) as z:
        z.writestr('[Content_Types].xml', types)
        z.writestr('_rels/.rels', relationships([(1, 'officeDocument', 'ppt/presentation.xml')]))
        z.writestr('ppt/presentation.xml', XML + f'<p:presentation {NS}><p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst><p:notesMasterIdLst><p:notesMasterId r:id="rId2"/></p:notesMasterIdLst><p:sldIdLst>{slide_ids}</p:sldIdLst><p:sldSz cx="{width}" cy="{height}"/><p:notesSz cx="6858000" cy="9144000"/></p:presentation>')
        z.writestr('ppt/_rels/presentation.xml.rels', relationships([(1, 'slideMaster', 'slideMasters/slideMaster1.xml'), (2, 'notesMaster', 'notesMasters/notesMaster1.xml')] + [(i+3, 'slide', f'slides/slide{i+1}.xml') for i in range(len(images))]))
        z.writestr('ppt/slideMasters/slideMaster1.xml', XML + f'<p:sldMaster {NS}><p:cSld><p:spTree>{GROUP}</p:spTree></p:cSld><p:clrMap {COLORS}/><p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rId1"/></p:sldLayoutIdLst><p:txStyles><p:titleStyle/><p:bodyStyle/><p:otherStyle/></p:txStyles></p:sldMaster>')
        z.writestr('ppt/slideMasters/_rels/slideMaster1.xml.rels', relationships([(1, 'slideLayout', '../slideLayouts/slideLayout1.xml'), (2, 'theme', '../theme/theme1.xml')]))
        z.writestr('ppt/slideLayouts/slideLayout1.xml', XML + f'<p:sldLayout {NS} type="blank" preserve="1"><p:cSld name="Blank"><p:spTree>{GROUP}</p:spTree></p:cSld><p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr></p:sldLayout>')
        z.writestr('ppt/slideLayouts/_rels/slideLayout1.xml.rels', relationships([(1, 'slideMaster', '../slideMasters/slideMaster1.xml')]))
        z.writestr('ppt/notesMasters/notesMaster1.xml', XML + f'<p:notesMaster {NS}><p:cSld><p:spTree>{GROUP}</p:spTree></p:cSld><p:clrMap {COLORS}/><p:notesStyle/></p:notesMaster>')
        z.writestr('ppt/notesMasters/_rels/notesMaster1.xml.rels', relationships([(1, 'theme', '../theme/theme1.xml')]))
        z.writestr('ppt/theme/theme1.xml', theme())
        for i, image in enumerate(images, 1):
            data = image.read_bytes()
            iw, ih = struct.unpack('>II', data[16:24])
            z.writestr(f'ppt/media/page{i}.png', data)
            z.writestr(f'ppt/slides/slide{i}.xml', picture_xml(width, height, iw, ih))
            z.writestr(f'ppt/slides/_rels/slide{i}.xml.rels', relationships([(1, 'slideLayout', '../slideLayouts/slideLayout1.xml'), (2, 'image', f'../media/page{i}.png'), (3, 'notesSlide', f'../notesSlides/notesSlide{i}.xml')]))
            z.writestr(f'ppt/notesSlides/notesSlide{i}.xml', notes_xml(notes.get(i-1, '')))
            z.writestr(f'ppt/notesSlides/_rels/notesSlide{i}.xml.rels', relationships([(1, 'notesMaster', '../notesMasters/notesMaster1.xml'), (2, 'slide', f'../slides/slide{i}.xml')]))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('pdf', nargs='?', type=Path, default=Path('slides.pdf'))
    parser.add_argument('output', nargs='?', type=Path, default=Path('main.pptx'))
    origin = parser.add_mutually_exclusive_group()
    origin.add_argument('--source', type=Path, help='Compile Typst and read its #speaker-note metadata')
    origin.add_argument('--notes', type=Path, help='Read existing pdfpc JSON metadata')
    parser.add_argument('--font-path', action='append', default=[])
    parser.add_argument('--input', action='append', default=[], help='Typst key=value input; repeat as needed')
    parser.add_argument('--width', type=int, default=1920, help='Maximum rendered image dimension in pixels')
    args = parser.parse_args()
    pdf, output = args.pdf.resolve(), args.output.resolve()
    metadata = {"pages": []}
    if args.source:
        source = args.source.resolve()
        pdf.parent.mkdir(parents=True, exist_ok=True)
        options = [value for p in args.font_path for value in ('--font-path', str(Path(p).resolve()))]
        options += [value for p in args.input for value in ('--input', p)]
        subprocess.run(['typst', 'compile', *options, source.name, str(pdf)], cwd=source.parent, check=True)
        result = subprocess.run(['typst', 'eval', 'query(<pdfpc-file>).first().value', '--in', source.name, '--format', 'json', *options], cwd=source.parent, check=True, capture_output=True, text=True)
        metadata = json.loads(result.stdout)
    elif args.notes:
        metadata = json.loads(args.notes.read_text())
    info = subprocess.check_output(['pdfinfo', str(pdf)], text=True, env={**os.environ, 'LC_ALL': 'C'})
    page_count = int(re.search(r'^Pages:\s+(\d+)', info, re.M).group(1))
    size = re.search(r'^Page size:\s+([\d.]+) x ([\d.]+)', info, re.M)
    width, height = (round(float(x) * 12700) for x in size.groups())
    notes = {}
    for page in metadata['pages']:
        idx = page['idx']
        if idx < 0 or idx >= page_count:
            parser.error(f'Note page index {idx} is outside this {page_count}-page PDF')
        if page.get('note'):
            notes[idx] = page['note']
    output.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='pdf2pptx-') as directory:
        prefix = Path(directory) / 'page'
        subprocess.run(['pdftoppm', '-scale-to', str(args.width), '-png', str(pdf), str(prefix)], check=True)
        images = sorted(Path(directory).glob('page-*.png'), key=lambda p: int(p.stem.split('-')[-1]))
        write_pptx(output, images, notes, width, height)
    print(f'{output}: {len(images)} slides, {len(notes)} with speaker notes')


if __name__ == '__main__':
    main()

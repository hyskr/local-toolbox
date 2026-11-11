#!/usr/bin/env python3
import json, sys
from pathlib import Path
import cv2
import numpy as np
from PIL import Image, ImageOps

config = json.loads(Path(sys.argv[1]).read_text())

def read_photo(path):
    im = ImageOps.exif_transpose(Image.open(path)).convert('RGB')
    im.thumbnail((2400, 2400))
    return im

def scan(im, points, mode, shadow):
    a = cv2.cvtColor(np.array(im), cv2.COLOR_RGB2BGR)
    if len(points) == 4:
        h,w = a.shape[:2]
        pts = np.float32([[x*w,y*h] for x,y in points])
        width = max(2, round(max(np.linalg.norm(pts[1]-pts[0]),np.linalg.norm(pts[2]-pts[3]))))
        height = max(2, round(max(np.linalg.norm(pts[3]-pts[0]),np.linalg.norm(pts[2]-pts[1]))))
        target = np.float32([[0,0],[width-1,0],[width-1,height-1],[0,height-1]])
        a = cv2.warpPerspective(a, cv2.getPerspectiveTransform(pts,target), (width,height))
    if shadow:
        channels=[]
        for plane in cv2.split(a):
            bg=cv2.medianBlur(cv2.dilate(plane, np.ones((7,7),np.uint8)),21)
            channels.append(cv2.divide(plane,bg,scale=255))
        a=cv2.merge(channels)
    if mode != '原色':
        gray=cv2.cvtColor(a,cv2.COLOR_BGR2GRAY)
        if mode == '黑白': gray=cv2.adaptiveThreshold(gray,255,cv2.ADAPTIVE_THRESH_GAUSSIAN_C,cv2.THRESH_BINARY,31,12)
        a=cv2.cvtColor(gray,cv2.COLOR_GRAY2BGR)
    return Image.fromarray(cv2.cvtColor(a,cv2.COLOR_BGR2RGB))

if config['action']=='scan':
    im=scan(read_photo(config['input']),config['points'],config['mode'],config['shadow'])
    im.save(config['output'],quality=94)
elif config['action']=='rotate':
    read_photo(config['input']).rotate(-90, expand=True).save(config['output'],quality=94)
elif config['action']=='export':
    scale=150/25.4
    pages=[]; page=Image.new('RGB',(round(210*scale),round(297*scale)),'white'); y=12; count=0
    for item in config['photos']:
        im=read_photo(item['path'])
        im=im.rotate(-90*item['rotation'],expand=True)
        width=min(config['width'],(273-config['gap'])*im.width/im.height)
        height=width*im.height/im.width
        if count and (config['onePage'] or y+height+config['gap']>285):
            pages.append(page); page=Image.new('RGB',page.size,'white'); y=12; count=0
        im=im.resize((max(1,round(width*scale)),max(1,round(height*scale))),Image.Resampling.LANCZOS)
        page.paste(im,(round(12*scale),round(y*scale))); y+=height+config['gap']; count+=1
    if count: pages.append(page)
    if not pages: raise ValueError('请先添加照片')
    pages[0].save(config['output'],save_all=True,append_images=pages[1:],resolution=150.0)

import AppKit
import UniformTypeIdentifiers

protocol FormatTarget { var title: String { get }; var ext: String { get }; var aliases: [String] { get } }
extension FormatTarget { var aliases: [String] { [] }; func matches(ext other: String) -> Bool { other == ext || aliases.contains(other) } }
enum CropWindow { static func run(image: CGImage, fileName: String) -> CGRect? { nil } }
enum FileKind: Equatable {
    case image, pdf, document, presentation, spreadsheet, video, audio, other
    init(url: URL) {
        switch url.pathExtension.lowercased() {
        case "pptx": self = .presentation; return
        case "xlsx": self = .spreadsheet; return
        default: break
        }
        let type = (try? url.resourceValues(forKeys: [.contentTypeKey]))?.contentType ?? UTType(filenameExtension: url.pathExtension.lowercased())
        guard let type else { self = .other; return }
        if type.conforms(to: .pdf) { self = .pdf }
        else if type.conforms(to: .image) { self = .image }
        else if type.conforms(to: .movie) { self = .video }
        else if type.conforms(to: .audio) { self = .audio }
        else if DocumentConverter.readableTypes.contains(where: { type.conforms(to: $0) }) { self = .document }
        else { self = .other }
    }
}
func convertFile(_ url: URL, target: String) throws -> [URL] {
    switch FileKind(url: url) {
    case .image:
        if target == "pdf" { return [try PDFConverter.makePDF(from: [url], near: url)] }
        guard let format=ImageConverter.targets.first(where: { $0.ext == target }), ImageConverter.canWrite(format.type) else { throw ConversionError.unsupported(target) }
        return [try ImageConverter.convert(url, to: format.type, ext: target)]
    case .pdf:
        if target == "txt" { return [try PDFConverter.toText(url)] }
        if target == "md" { return [try PDFConverter.toMarkdown(url)] }
        if target == "docx" || target == "rtf", let format=DocumentConverter.targets.first(where: {$0.ext == target}) { return [try PDFConverter.toDocument(url, target: format)] }
        guard let format=ImageConverter.targets.first(where: { $0.ext == target }), ImageConverter.canWrite(format.type) else { throw ConversionError.unsupported(target) }
        return try PDFConverter.toImages(url, type: format.type, ext: target)
    case .document:
        guard let format=DocumentConverter.targets.first(where: {$0.ext == target}) else {throw ConversionError.unsupported(target)}
        return [try DocumentConverter.convert(url, to: format)]
    case .presentation:
        if target == "md" { return [try PresentationConverter.toMarkdown(url)] }
        if target == "pdf" { return [try PresentationConverter.toPDF(url)] }
        throw ConversionError.unsupported(target)
    case .spreadsheet:
        if target == "md" { return [try SpreadsheetConverter.toMarkdown(url)] }
        if target == "pdf" { return [try SpreadsheetConverter.toPDF(url)] }
        throw ConversionError.unsupported(target)
    case .video:
        guard let format=MediaConverter.videoTargets.first(where: {$0.ext == target}) else {throw ConversionError.unsupported(target)}
        return [try MediaConverter.convertVideo(url, to: format)]
    case .audio:
        guard let format=MediaConverter.audioTargets.first(where: {$0.ext == target}) else {throw ConversionError.unsupported(target)}
        return [try MediaConverter.convertAudio(url, to: format)]
    case .other: throw ConversionError.unsupported(url.pathExtension)
    }
}
let args = CommandLine.arguments
if args.count < 3 { fputs("请选择输入文件和输出格式\n", stderr); exit(2) }
let outputFormat = args[1]
func emitEvent(_ event: [String: Any]) {
    if let data = try? JSONSerialization.data(withJSONObject: event) { print(String(decoding:data,as:UTF8.self)); fflush(stdout) }
}
for path in args.dropFirst(2) {
    emitEvent(["path":path,"state":"running"])
    do {
        let outputs = try convertFile(URL(fileURLWithPath:path),target:outputFormat).map(\.path)
        emitEvent(["path":path,"state":"done","outputs":outputs])
    } catch {
        emitEvent(["path":path,"state":"failed","error":error.localizedDescription])
    }
}

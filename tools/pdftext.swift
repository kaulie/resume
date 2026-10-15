import Foundation
import PDFKit

// 极简 PDF 文本抽取器：用 macOS 原生 PDFKit 抽取整篇 PDF 的文本（含中文/CID 字体）。
// 由 Go 后端在解析 PDF 时调用（比纯 Go 的简易解析器对中文 PDF 更可靠）。
guard CommandLine.arguments.count > 1 else {
    FileHandle.standardError.write("usage: pdftext <file.pdf>\n".data(using: .utf8)!)
    exit(2)
}

let path = CommandLine.arguments[1]
guard let doc = PDFDocument(url: URL(fileURLWithPath: path)) else {
    FileHandle.standardError.write("pdftext: cannot open pdf\n".data(using: .utf8)!)
    exit(1)
}

var out = ""
for i in 0..<doc.pageCount {
    if let page = doc.page(at: i), let text = page.string {
        out += text
        if !text.hasSuffix("\n") { out += "\n" }
    }
}
print(out, terminator: "")

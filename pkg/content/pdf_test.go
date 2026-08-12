/*
 * 文件作用：pkg/content PDF Parser 的单元测试 -- 程序化最小 PDF 夹具的文本提取与非法输入报错。
 * 创建日期：2026-08-12
 * 修改日期：2026-08-12
 */
package content

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"rove/pkg/document"
)

// writePDFFixture 生成含 "Hello Rove" 文本的最小合法 PDF（xref 偏移程序化计算）。
func writePDFFixture(t *testing.T) []byte {

	t.Helper()
	stream := "BT /F1 24 Tf 72 720 Td (Hello Rove) Tj ET\n"
	objects := []string{
		"1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n",
		"2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n",
		"3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n",
		fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(stream), stream),
		"5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n",
	}
	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects))
	for index, obj := range objects {
		offsets[index] = buf.Len()
		buf.WriteString(obj)
	}
	xref := buf.Len()
	buf.WriteString("xref\n0 " + strconv.Itoa(len(objects)+1) + "\n")
	buf.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets {
		buf.WriteString(fmt.Sprintf("%010d 00000 n \n", offset))
	}
	buf.WriteString("trailer\n<< /Size " + strconv.Itoa(len(objects)+1) + " /Root 1 0 R >>\n")
	buf.WriteString("startxref\n" + strconv.Itoa(xref) + "\n%%EOF")
	return buf.Bytes()
}

func TestPDFParser(t *testing.T) {

	parser, err := DefaultRegistry().For("application/pdf")
	require.NoError(t, err, "pdf must be registered")

	raw := &document.RawDocument{ID: "pdf-1", URL: "https://x.com/doc.pdf", ContentType: "application/pdf", Body: writePDFFixture(t)}
	parsed, err := parser.Parse(context.Background(), raw)
	require.NoError(t, err)
	require.Contains(t, parsed.Content, "Hello Rove")
}

func TestPDFParserInvalid(t *testing.T) {

	parser, _ := DefaultRegistry().For("application/pdf")
	raw := &document.RawDocument{ID: "pdf-2", URL: "https://x.com/bad.pdf", ContentType: "application/pdf", Body: []byte("not a pdf")}
	_, err := parser.Parse(context.Background(), raw)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pdf")
}

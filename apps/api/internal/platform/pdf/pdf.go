package pdf

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

func esc(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "(", "\\(")
	return strings.ReplaceAll(s, ")", "\\)")
}
func Build(title string, lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT /F1 18 Tf 50 790 Td (" + esc(title) + ") Tj /F1 10 Tf 0 -28 Td ")
	for _, l := range lines {
		content.WriteString("(" + esc(l) + ") Tj 0 -16 Td ")
	}
	content.WriteString("ET")
	stream := content.String()
	objs := []string{"<< /Type /Catalog /Pages 2 0 R >>", "<< /Type /Pages /Kids [3 0 R] /Count 1 >>", "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Resources << /Font << /F1 5 0 R >> >> /Contents 4 0 R >>", fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream), "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>"}
	var out bytes.Buffer
	out.WriteString("%PDF-1.4\n")
	offset := []int{0}
	for i, o := range objs {
		offset = append(offset, out.Len())
		out.WriteString(strconv.Itoa(i+1) + " 0 obj\n" + o + "\nendobj\n")
	}
	xref := out.Len()
	out.WriteString(fmt.Sprintf("xref\n0 %d\n0000000000 65535 f \n", len(objs)+1))
	for i := 1; i < len(offset); i++ {
		out.WriteString(fmt.Sprintf("%010d 00000 n \n", offset[i]))
	}
	out.WriteString(fmt.Sprintf("trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objs)+1, xref))
	return out.Bytes()
}

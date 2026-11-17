package cloud.eka_dev.eka_hub.feature.chat

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class GeminiMarkdownRendererTest {

    @Test
    fun testParseMarkdownBlocks_codeBlock() {
        val markdown = """
            Berikut adalah fungsi di Go:
            ```go
            func add(a int, b int) int {
                return a + b
            }
            ```
            Penjelasan fungsi di atas.
        """.trimIndent()

        val blocks = parseMarkdownBlocks(markdown)

        assertEquals(3, blocks.size)
        assertTrue(blocks[0] is MarkdownBlock.Paragraph)
        assertEquals("Berikut adalah fungsi di Go:", (blocks[0] as MarkdownBlock.Paragraph).text)

        assertTrue(blocks[1] is MarkdownBlock.CodeBlock)
        val codeBlock = blocks[1] as MarkdownBlock.CodeBlock
        assertEquals("go", codeBlock.language)
        assertTrue(codeBlock.code.contains("func add(a int, b int) int"))
        assertTrue(codeBlock.code.contains("return a + b"))

        assertTrue(blocks[2] is MarkdownBlock.Paragraph)
        assertEquals("Penjelasan fungsi di atas.", (blocks[2] as MarkdownBlock.Paragraph).text)
    }

    @Test
    fun testParseMarkdownBlocks_headingsAndBullets() {
        val markdown = """
            # Judul Utama
            ## Sub Judul
            - Poin pertama
            - Poin kedua
            1. Langkah satu
            2. Langkah dua
        """.trimIndent()

        val blocks = parseMarkdownBlocks(markdown)

        assertEquals(6, blocks.size)
        assertTrue(blocks[0] is MarkdownBlock.Heading)
        assertEquals(1, (blocks[0] as MarkdownBlock.Heading).level)
        assertEquals("Judul Utama", (blocks[0] as MarkdownBlock.Heading).text)

        assertTrue(blocks[1] is MarkdownBlock.Heading)
        assertEquals(2, (blocks[1] as MarkdownBlock.Heading).level)
        assertEquals("Sub Judul", (blocks[1] as MarkdownBlock.Heading).text)

        assertTrue(blocks[2] is MarkdownBlock.BulletItem)
        assertEquals("Poin pertama", (blocks[2] as MarkdownBlock.BulletItem).text)

        assertTrue(blocks[3] is MarkdownBlock.BulletItem)
        assertEquals("Poin kedua", (blocks[3] as MarkdownBlock.BulletItem).text)

        assertTrue(blocks[4] is MarkdownBlock.NumberedItem)
        assertEquals("1", (blocks[4] as MarkdownBlock.NumberedItem).number)
        assertEquals("Langkah satu", (blocks[4] as MarkdownBlock.NumberedItem).text)

        assertTrue(blocks[5] is MarkdownBlock.NumberedItem)
        assertEquals("2", (blocks[5] as MarkdownBlock.NumberedItem).number)
        assertEquals("Langkah dua", (blocks[5] as MarkdownBlock.NumberedItem).text)
    }

    @Test
    fun testHighlightSyntax_returnsAnnotatedString() {
        val code = """
            // Komentar
            func main() {
                val x = 123
                val s = "Halo dunia"
            }
        """.trimIndent()

        val highlighted = highlightSyntax(code, "go")
        assertEquals(code, highlighted.text)
    }
}

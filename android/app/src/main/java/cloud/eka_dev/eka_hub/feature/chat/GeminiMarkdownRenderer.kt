package cloud.eka_dev.eka_hub.feature.chat

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import cloud.eka_dev.eka_hub.core.designsystem.*
import kotlinx.coroutines.delay
import java.util.regex.Pattern

/**
 * Structured block parsed from markdown text.
 */
sealed class MarkdownBlock {
    data class CodeBlock(val language: String, val code: String) : MarkdownBlock()
    data class Heading(val level: Int, val text: String) : MarkdownBlock()
    data class BulletItem(val text: String) : MarkdownBlock()
    data class NumberedItem(val number: String, val text: String) : MarkdownBlock()
    data class Paragraph(val text: String) : MarkdownBlock()
}

/**
 * Gemini-Grade Markdown Renderer for Jetpack Compose.
 */
@Composable
fun GeminiMarkdownContent(
    content: String,
    modifier: Modifier = Modifier,
    textColor: Color = Color(0xFFF1F5F9)
) {
    val blocks = remember(content) { parseMarkdownBlocks(content) }

    Column(
        modifier = modifier,
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        blocks.forEach { block ->
            when (block) {
                is MarkdownBlock.CodeBlock -> {
                    GeminiCodeCard(
                        language = block.language,
                        code = block.code
                    )
                }
                is MarkdownBlock.Heading -> {
                    val (fontSize, fontWeight) = when (block.level) {
                        1 -> 18.sp to FontWeight.Bold
                        2 -> 16.sp to FontWeight.SemiBold
                        else -> 14.sp to FontWeight.SemiBold
                    }
                    Text(
                        text = buildInlineMarkdown(block.text, textColor),
                        fontSize = fontSize,
                        fontWeight = fontWeight,
                        color = Color.White,
                        lineHeight = (fontSize.value * 1.4f).sp,
                        modifier = Modifier.padding(top = 4.dp, bottom = 2.dp)
                    )
                }
                is MarkdownBlock.BulletItem -> {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(start = 4.dp),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalAlignment = Alignment.Top
                    ) {
                        Text(
                            text = "•",
                            color = HubCyan,
                            fontSize = 15.sp,
                            fontWeight = FontWeight.Bold,
                            lineHeight = 21.sp
                        )
                        Text(
                            text = buildInlineMarkdown(block.text, textColor),
                            fontSize = 14.sp,
                            color = textColor,
                            lineHeight = 21.sp,
                            modifier = Modifier.weight(1f)
                        )
                    }
                }
                is MarkdownBlock.NumberedItem -> {
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(start = 4.dp),
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalAlignment = Alignment.Top
                    ) {
                        Text(
                            text = "${block.number}.",
                            color = HubIndigo,
                            fontSize = 13.sp,
                            fontWeight = FontWeight.Bold,
                            lineHeight = 21.sp
                        )
                        Text(
                            text = buildInlineMarkdown(block.text, textColor),
                            fontSize = 14.sp,
                            color = textColor,
                            lineHeight = 21.sp,
                            modifier = Modifier.weight(1f)
                        )
                    }
                }
                is MarkdownBlock.Paragraph -> {
                    Text(
                        text = buildInlineMarkdown(block.text, textColor),
                        fontSize = 14.sp,
                        color = textColor,
                        lineHeight = 21.sp
                    )
                }
            }
        }
    }
}

/**
 * Dedicated Code Block Card with syntax highlighting, language badge, and copy button.
 */
@Composable
fun GeminiCodeCard(
    language: String,
    code: String,
    modifier: Modifier = Modifier
) {
    val clipboardManager = LocalClipboardManager.current
    var copied by remember { mutableStateOf(false) }

    LaunchedEffect(copied) {
        if (copied) {
            delay(2000)
            copied = false
        }
    }

    val displayLang = remember(language) {
        if (language.isBlank()) "CODE" else language.uppercase()
    }

    val highlightedCode = remember(code, language) {
        highlightSyntax(code, language.lowercase())
    }

    Column(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(Color(0xFF0B0E17))
            .border(1.dp, Color(0xFF232838), RoundedCornerShape(12.dp))
    ) {
        // Code Block Header Bar
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .background(Color(0xFF131722))
                .padding(horizontal = 14.dp, vertical = 8.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = displayLang,
                fontFamily = FontFamily.Monospace,
                fontSize = 11.sp,
                fontWeight = FontWeight.SemiBold,
                color = Color(0xFF94A3B8)
            )

            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(6.dp))
                    .clickable {
                        clipboardManager.setText(AnnotatedString(code))
                        copied = true
                    }
                    .padding(horizontal = 8.dp, vertical = 4.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                AnimatedVisibility(visible = copied, enter = fadeIn(), exit = fadeOut()) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.Check,
                            contentDescription = "Tersalin",
                            tint = Color(0xFF4ADE80),
                            modifier = Modifier.size(13.dp)
                        )
                        Text(
                            text = "Tersalin!",
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Medium,
                            color = Color(0xFF4ADE80)
                        )
                    }
                }
                AnimatedVisibility(visible = !copied, enter = fadeIn(), exit = fadeOut()) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(4.dp)
                    ) {
                        Icon(
                            imageVector = Icons.Default.ContentCopy,
                            contentDescription = "Salin",
                            tint = Color(0xFF94A3B8),
                            modifier = Modifier.size(13.dp)
                        )
                        Text(
                            text = "Salin kode",
                            fontSize = 11.sp,
                            fontWeight = FontWeight.Medium,
                            color = Color(0xFF94A3B8)
                        )
                    }
                }
            }
        }

        // Code Body with horizontal scrolling
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState())
                .padding(horizontal = 14.dp, vertical = 12.dp)
        ) {
            Text(
                text = highlightedCode,
                fontFamily = FontFamily.Monospace,
                fontSize = 13.sp,
                lineHeight = 19.sp,
                color = Color(0xFFF1F5F9)
            )
        }
    }
}

/**
 * Parses markdown text into ordered structured blocks.
 */
fun parseMarkdownBlocks(text: String): List<MarkdownBlock> {
    val blocks = mutableListOf<MarkdownBlock>()
    val lines = text.lines()
    var i = 0

    while (i < lines.size) {
        val line = lines[i]
        val trimmed = line.trim()

        // Check for fenced code block: ```language
        if (trimmed.startsWith("```")) {
            val lang = trimmed.removePrefix("```").trim()
            val codeLines = mutableListOf<String>()
            i++
            while (i < lines.size && !lines[i].trim().startsWith("```")) {
                codeLines.add(lines[i])
                i++
            }
            if (i < lines.size && lines[i].trim().startsWith("```")) {
                i++ // consume closing ```
            }
            blocks.add(MarkdownBlock.CodeBlock(language = lang, code = codeLines.joinToString("\n")))
            continue
        }

        // Heading: # H1, ## H2, ### H3
        if (trimmed.startsWith("#")) {
            var level = 0
            while (level < trimmed.length && trimmed[level] == '#') {
                level++
            }
            val headingText = trimmed.substring(level).trim()
            if (headingText.isNotEmpty()) {
                blocks.add(MarkdownBlock.Heading(level = level.coerceIn(1, 4), text = headingText))
                i++
                continue
            }
        }

        // Bullet point: - item, * item
        if (trimmed.startsWith("- ") || trimmed.startsWith("* ")) {
            blocks.add(MarkdownBlock.BulletItem(text = trimmed.substring(2).trim()))
            i++
            continue
        }

        // Numbered list: 1. item, 2. item
        val numMatcher = Pattern.compile("^(\\d+)\\.\\s+(.*)$").matcher(trimmed)
        if (numMatcher.matches()) {
            val num = numMatcher.group(1) ?: "1"
            val itemText = numMatcher.group(2) ?: ""
            blocks.add(MarkdownBlock.NumberedItem(number = num, text = itemText))
            i++
            continue
        }

        // Paragraph line
        if (trimmed.isNotEmpty()) {
            blocks.add(MarkdownBlock.Paragraph(text = line))
        }
        i++
    }

    return blocks
}

/**
 * Builds formatted AnnotatedString supporting inline markdown:
 * - **bold**
 * - *italic*
 * - `inline code`
 */
fun buildInlineMarkdown(text: String, defaultColor: Color): AnnotatedString {
    return buildAnnotatedString {
        var cursor = 0
        val length = text.length

        while (cursor < length) {
            // Inline code: `code`
            if (text[cursor] == '`') {
                val end = text.indexOf('`', cursor + 1)
                if (end != -1) {
                    val codeSnippet = text.substring(cursor + 1, end)
                    pushStyle(
                        SpanStyle(
                            fontFamily = FontFamily.Monospace,
                            color = Color(0xFF93C5FD),
                            background = Color(0xFF1E2433),
                            fontSize = 12.sp,
                            fontWeight = FontWeight.Medium
                        )
                    )
                    append(" $codeSnippet ")
                    pop()
                    cursor = end + 1
                    continue
                }
            }

            // Bold: **text**
            if (cursor + 1 < length && text[cursor] == '*' && text[cursor + 1] == '*') {
                val end = text.indexOf("**", cursor + 2)
                if (end != -1) {
                    val boldSnippet = text.substring(cursor + 2, end)
                    pushStyle(SpanStyle(fontWeight = FontWeight.Bold, color = Color.White))
                    append(boldSnippet)
                    pop()
                    cursor = end + 2
                    continue
                }
            }

            // Italic: *text* (single asterisk)
            if (text[cursor] == '*') {
                val end = text.indexOf('*', cursor + 1)
                if (end != -1 && end > cursor + 1) {
                    val italicSnippet = text.substring(cursor + 1, end)
                    pushStyle(SpanStyle(fontStyle = FontStyle.Italic, color = defaultColor))
                    append(italicSnippet)
                    pop()
                    cursor = end + 1
                    continue
                }
            }

            append(text[cursor])
            cursor++
        }
    }
}

/**
 * Lexical Syntax Highlighter for Kotlin, Swift, Go, Python, SQL, JSON, Bash, JS/TS.
 */
fun highlightSyntax(code: String, language: String): AnnotatedString {
    return buildAnnotatedString {
        append(code)

        val keywordColor = Color(0xFFC084FC) // Lilac / Purple
        val stringColor = Color(0xFF86EFAC)  // Mint Green
        val numberColor = Color(0xFFFDBA74)  // Amber / Orange
        val commentColor = Color(0xFF94A3B8) // Slate Gray
        val typeColor = Color(0xFF38BDF8)    // Sky Blue

        // Keyword dictionary across major languages
        val keywords = listOf(
            "fun", "val", "var", "class", "interface", "object", "package", "import", "return", "if", "else", "when",
            "while", "for", "in", "is", "as", "try", "catch", "finally", "throw", "override", "private", "public",
            "func", "let", "struct", "enum", "protocol", "extension", "guard", "switch", "case", "default", "defer",
            "type", "const", "go", "select", "chan", "range", "map",
            "def", "elif", "except", "with", "lambda", "yield", "pass", "from",
            "SELECT", "FROM", "WHERE", "JOIN", "LEFT", "RIGHT", "INNER", "GROUP", "BY", "ORDER", "LIMIT", "INSERT", "INTO", "VALUES", "UPDATE", "SET", "DELETE", "AND", "OR", "NOT",
            "function", "async", "await", "export", "const"
        )

        val types = listOf(
            "String", "Int", "Long", "Float", "Double", "Boolean", "Unit", "Any", "List", "Map", "Set",
            "int", "int64", "string", "bool", "float64", "error", "byte",
            "true", "false", "nil", "null", "None"
        )

        // Highlight Keywords
        for (kw in keywords) {
            val pattern = Pattern.compile("\\b$kw\\b")
            val matcher = pattern.matcher(code)
            while (matcher.find()) {
                addStyle(SpanStyle(color = keywordColor, fontWeight = FontWeight.Bold), matcher.start(), matcher.end())
            }
        }

        // Highlight Types
        for (tp in types) {
            val pattern = Pattern.compile("\\b$tp\\b")
            val matcher = pattern.matcher(code)
            while (matcher.find()) {
                addStyle(SpanStyle(color = typeColor), matcher.start(), matcher.end())
            }
        }

        // Highlight Strings: "..." or '...'
        val strPattern = Pattern.compile("\"(\\\\.|[^\"])*\"|'(\\\\.|[^'])*'")
        val strMatcher = strPattern.matcher(code)
        while (strMatcher.find()) {
            addStyle(SpanStyle(color = stringColor), strMatcher.start(), strMatcher.end())
        }

        // Highlight Numbers: integers & floats
        val numPattern = Pattern.compile("\\b\\d+(\\.\\d+)?\\b")
        val numMatcher = numPattern.matcher(code)
        while (numMatcher.find()) {
            addStyle(SpanStyle(color = numberColor), numMatcher.start(), numMatcher.end())
        }

        // Highlight Comments: // ... or # ... or /* ... */
        val commentPattern = Pattern.compile("//.*|#.*|/\\*[\\s\\S]*?\\*/")
        val commentMatcher = commentPattern.matcher(code)
        while (commentMatcher.find()) {
            addStyle(SpanStyle(color = commentColor, fontStyle = FontStyle.Italic), commentMatcher.start(), commentMatcher.end())
        }
    }
}

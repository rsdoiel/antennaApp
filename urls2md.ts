// Function to parse input text and generate markdown
async function transformToMarkdown(filePath: string) {
    // Read the input file content
    const content = await Deno.readTextFile(filePath);
    const lines = content.split('\n');

    let markdownLines: string[] = [];

    // Process each line to extract URL and description
    for (const line of lines) {
        if (line.trim().startsWith('http')) {
            // Split on quote marks to separate URL and description
            const parts = line.split('"');
            if (parts.length >= 2) {
                const url = parts[0].trim();
                const description = parts[1].trim().replace(/^~/g,'');
                // Create markdown line
                markdownLines.push(`- [${description}](${url})`);
            }
        } else if (line.trim() === "#") {
            markdownLines.push('');
        } else {
            markdownLines.push(line);
        }
    }

    // Join markdown lines into a single string
    const markdownContent = markdownLines.join('\n');
    console.log(markdownContent);
}

async function main() {
    // Get file path from command line arguments
    const filePath = Deno.args[0];
    if (!filePath) {
        console.error("Please provide a file path as an argument.");
        Deno.exit(1);
    }

    // Call the transform function
    await transformToMarkdown(filePath);
}

if (import.meta.main) {
    await main();
}
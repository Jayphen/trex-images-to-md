## What is this?

I am currently in my _note taking optimisation_ era, and have found that I rather enjoy taking notes on paper.

However, I also want to be able to index my notes and feed them in to [Silverbullet](https://silverbullet.md/), which I run on a local server.

So, my workflow is to scan my notes daily to jpeg, and use [TRex](https://github.com/amebalabs/TRex/tree/main) to OCR them. TRex has a CLI tool, which I am utilising here to pull the text out of the images. Because my handwriting is not perfect, OCRing does not usually produce a very clean output. To correct this, I send the output to GPT4 and ask it politely to correct typos and format the result as markdown.

## How to use it?

MacOS only.

Install TRex and add the cli tool (/Applications/TRex.app/Contents/MacOS/cli/trex) to your path.

Add your key to the OPENAI_API_KEY environment variable in your shell.

`go run main.go path/to/your/image` will create an .md file with the same name as your image.

## Todo

- [x] Configurable system message
- [ ] Configurable dest directory
- [ ] Cleanup when I learn how to write Go
- [x] Use goroutines for multiple files as input
- [ ] A binary release maybe
- [ ] Embed image by adding path to the .md, move it along with the .md (configurable)

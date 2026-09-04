import MathJax from 'mathjax';
import fs from 'node:fs';
import process from 'node:process';

await MathJax.init({
    loader: {load: ['input/tex', 'output/svg']}
});

// Assumed param values:
// 0 = Node runtime path
// 1 = Path of this script
// 2 = TeX content
// 3 = Output file path
const texString = process.argv[2];
const outputFile = process.argv[3];

console.log("Input", texString)
console.log("Output", outputFile)

const mathjaxContainerNode = await MathJax.tex2svgPromise(texString, {display: true});
const svgNode = mathjaxContainerNode.children[0];
const svgContent = MathJax.startup.adaptor.serializeXML(svgNode)
console.log("SVG content", svgContent)

fs.writeFile(outputFile, svgContent, err => {
    console.log("Writing error", err)
    if (err) {
        console.error("Could not write to output file " + outputFile, err);
    } else {
        // file written successfully
    }
});
package parser

const FILE_PREFIXES = "Datei|File|Bild|Image|Media"
const IMAGE_REGEX_PATTERN = `\[\[((` + FILE_PREFIXES + `):([^|^\]]*))(\|([^\]]*))?]]`

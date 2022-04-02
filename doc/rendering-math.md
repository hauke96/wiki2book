TeX snippets embedded with the `<math>...</math>` tags are rendered using the Wikipedia API.
The rendering takes place in `api.go` and is called by `html.go` when expanding the math-token.

# The Wikipedia API

This API consists of two URLs:

1. The Tex checking URL to create an resource token from the math string: `https://wikimedia.org/api/rest_v1/media/math/check/tex`
2. The rendering URLs to actually get the images (see below for what `HASH` is):
   1. For the SVG: `https://wikimedia.org/api/rest_v1/media/math/render/svg/HASH`
   1. For the PNG: `https://wikimedia.org/api/rest_v1/media/math/render/png/HASH`

To turn the math string into an SVG, the following two steps are needed (one per URL):

## 1. Checking the TeX → Getting image hash

You talk to this URL with an POST request havin `application/x-www-form-urlencoded` as `contentTape` and the string `q=<your-tex-math-string>` as body.

The response doesn't matter but the `x-resource-location` header of that response does.
It contains a hash value (described as `HASH` above; I like to call it *SVG-token*) you need to get the SVG or PNG from the image rendering API (see next step).

## 2. Rendering the Math to an image

Using the hash/SVG-token from above, you can just make a GET request to `https://wikimedia.org/api/rest_v1/media/math/render/svg/<your-svg-token>` to download the SVG.
Just replace `.../svg/...` by `.../png/...` to get the SVG as PNG.

In the eBook PNGs are not as pretty as SVGs but a safer option if you have issue with rendering SVGs.

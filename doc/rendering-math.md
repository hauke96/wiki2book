TeX snippets embedded with the `<math>...</math>` tags are rendered using the Wikipedia API.
The rendering takes place in `api.go` and is called by `html.go` when expanding the math-token.

# The Wikipedia API

This API consists of two URLs:

1. The Tex checking URL to create an resource token from the math string: `https://wikimedia.org/api/rest_v1/media/math/check/tex`
2. The rendering URL to actually get the SVG: `https://wikimedia.org/api/rest_v1/media/math/render/svg/`

To turn the math string into an SVG, the following two steps are needed (one per URL):

## 1. Checking the TeX → Getting SVG-token

You talk to this URL with an POST request havin `application/x-www-form-urlencoded` as `contentTape` and the string `q=<your-tex-math-string>` as body.

The response doesn't matter but the `x-resource-location` header of that response does.
It contains a hash value (I like to call it *SVG-token*) you need to get the SVG from the image rendering API (see next step).

## 2. Rendering the SVG

Using the SVG-token from above, you can just make a GET request to `https://wikimedia.org/api/rest_v1/media/math/render/svg/<yout-svg-token>` to download the SVG.

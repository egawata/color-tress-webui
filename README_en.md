# color-tress-webui : Color Tress Layer Generator

[Japanese Doc](README.md)

## Overview

This is a tool for easily generating layers for color tressing.

## Build

To run it as is, you need [tinygo](https://tinygo.org/).
If you want to use the official Go, edit the commented-out parts as appropriate.

~~~sh
scripts/build.sh
~~~

## Execution

If an HTTP server is available, publish the contents of the `./build` directory (or its copied directory).
A simple HTTP server is also provided.  Run the following:

~~~sh
go run localserver/run_server.go
~~~

## Usage

Detailed usage is also available at https://egawata.tokyo/color-tress.

- Prepare an image with the line art layer hidden (PNG format recommended).
- Open `colortress.html`.
- Click `Select Image` button and choose the prepared image.
- Click `Generate` button.
- Once the output image is generated, click `Download` button to save it.
- Load this image into your paint tool and place it just above the line art layer.
- Set it to clip to the layer below.
- Adjust the opacity as needed.

## Web Version

A web version of this tool is also available at https://egawata.tokyo/color-tress/colortress.html.

## Feedback and Contributions

Your feedback or pull requests are highly appreciated.

## License

Licensed under the Apache 2.0 license. Copyright (c) 2025 by egawata

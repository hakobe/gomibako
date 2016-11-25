# Gomibako

A requestb.in clone writtern in Go. (hobby product)

## Usage

```sh
$ npm install -g yarn # If you have not installed yarn
$ make
$ ./gomibako --port=8000
```

And access to http://localhost:8000 .

You can also build it for a Linux environment like below.

```
$ make clean # If you need
$ GOOS=linux GOARCH=amd64 make
```

## Description

Gomibako is a mini web app to inspect HTTP requests to it (like [requestb.in](http://requestb.in/)).

- Gives you a URL that collect requests to it
- Let you inspect requests in **real time** (using Sever Sent Event)

## License

[MIT](./LICENSE)

## Author

[hakobe](http://github.com/hakobe)

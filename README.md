# Šparovec

Šparovec (slovenian slang for piggy bank) is a very simple web app for managing personal finances. Simplicity is the core idea behind this project,
therefore its written in go + htmx + sqlite. This results in a single executable that can be hosted anywhere you want.

## Quickstart

1. Run:

```sh
pnpm install
```

to install the [taiwlindcss](https://tailwindcss.com/) dependency used to generate the `css` styles.

2. Run:

```sh
make build
```

to build the executable.

3. Add user with

```sh
./sparovec create-user <username> <password>
```

4. Start the server with

```sh
./sparovec serve
```

## Development

The following tools are required for development:

- `golang`: main language of the project,
- [templ](https://templ.guide/): templating library for go,
- `pnpm`: used for running the tailwindcss tool for building the `css` styles,
- [air](https://github.com/cosmtrek/air): hot reloading during development (optional)

1. Run:

```sh
pnpm tailwindcss -i ./assets/_global.css -o ./assets/global.css --watch
```

to build the `css` styles live.

2. Run:

```sh
air
```

to hot reload the server during development.

3. Open the project in you favorite editor :)

## License

[GNU General Public License v3.0](LICENSE)


# Vie

> **Work in Progress**: This project is under active development. The API is
> not stable and may change at any time.

Vie (French: /vi/, meaning "life") is a lightweight toolkit for code
generation, designed to make creating and reusing templates simple and
efficient.

You can templatize all boilerplate in your project: new Java classes, API
handlers, React components, and more.

For example, when creating a React component:

- If you create just one file, it's fine.
- But if you create many files (e.g., tests, `Component.props.ts`, or additional folders),
  you can generate a **file structure** with general parameters and
  boilerplate in all files using a single command.

## Quick start

Create a `.vie/` directory. All templates will be stored there. Inside it,
create a `my-template/` directory. It represents a single template, and
all files inside it will be used for rendering output.

Create `.vie/my-template/hello.txt.vie` with the following content:

```vie
Greetings to {{ name }} from Vie!
```

Render the template using the `new` command:

```bash
vie new my-template src/ name=HappyUser
```

This generates a file `src/hello.txt` with the following content:

<!-- TODO: add link to language reference -->

```txt
Greetings to HappyUser from Vie!
```

## Installation

```sh
go install github.com/vietmpl/vie@latest
```

<!-- TODO: add installation via package managers -->

## Licinse

`vie` is distributed under the terms of the [MIT License](./LICINSE).

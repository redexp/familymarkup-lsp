# FamilyMarkup Language Server

[![Go Reference](https://pkg.go.dev/badge/github.com/redexp/familymarkup-lsp.svg)](https://pkg.go.dev/github.com/redexp/familymarkup-lsp)

## Features

- [x] SemanticTokens
  - [x] For full document
  - [ ] For range of a document
  - [x] Delta
- [x] Completion of:
  - Names filtered by surname (in case when cursor before surname)
  - Surnames filtered by name (in case when you wrote the name and start writing surname)
  - Names or Surnames in all other cases
- [x] Jump to Definition (usually Ctrl + Click) of any name or surname
- [x] Find All References of names or surnames 
- [x] "Go to Type Definition" — jump to a Markdown file by person's name in a file path like `Potter/Harry.md` or `Potter/Harry/index.md`
- [x] Hover hints. Show highlighted a hint about person in format like `Name - child of Name + Name`
- [x] DocumentHighlight — highlight of all references of currently focused name or surname in the current file
- [x] Rename
- [x] Folding
- [x] CodeAction
  - [x] QuickFix for "Unknown family" error
  - [x] QuickFix for "An unobvious name" warning
- [x] Symbol
  - [x] For current document - in editor could be shown in file path toolbar as surname and name of currently focused name like `Potter.family * Potter * Harry`
  - [x] For workspace - helpful to find any person from any place like in vscode by running command `#HarPot` will show all people which name starts with `Har` and surname with `Pot`
- [x] Tree view - helpful to build family tree like
    ```
    Weasley
    └── Arthur + Molly?
        ├── Fred
        ├── George
        ├── Ronald
        └── girl?
    ```

## Configurations

### Language

Language of error messages, hints and so on

- [x] English
- [x] Українська
- [x] Русский

## Ideas / New Features / TODO

Feel free to open an issue with your idea how to improve code editing or navigation.

- [ ] If person has changed surname then his name can be used in that family without origin surname.
- [ ] Quick fix for family name to detach family and all it members into a separate file
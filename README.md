# FamilyMarkup Language Server

[![Go Reference](https://pkg.go.dev/badge/github.com/redexp/familymarkup-lsp.svg)](https://pkg.go.dev/github.com/redexp/familymarkup-lsp)

## Features

- [x] SemanticTokens
  - [x] For full document
  - [ ] For range of document
  - [x] Delta
- [x] Completion of:
  - Names filtered by surname (in case when cursor before surname)
  - Surnames filtered by name (in case when you wrote the name and start writing surname)
  - Names or Surnames in all other cases
- [x] Jump to Definition (usually Ctrl + Click) of any name or surname
- [x] Find All References of names or surnames 
- [x] "Go to Type Definition" - jump to markdown file by person's name in file path like `families/Potter/Harry.md`
- [x] Hover hints. Show highlighted hint about person in format like `Name - child of Name + Name`
- [x] DocumentHighlight - highlight of all references of currently focused name or surname in current file
- [x] Rename
- [x] Folding
- [x] CodeAction
  - [x] QuickFix for "Unknown family" error
  - [x] QuickFix for "An unobvious name" warning
- [x] Symbol
  - [x] For current document - in editor could be shown in file path toolbar as surname and name of currently focused name like `families/Potter.family * Potter * Harry`
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

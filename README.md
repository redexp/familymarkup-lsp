# FamilyMarkup Language Server

[![Go Reference](https://pkg.go.dev/badge/github.com/redexp/familymarkup-lsp.svg)](https://pkg.go.dev/github.com/redexp/familymarkup-lsp)

## Features

- [x] SemanticTokens
  - [x] For full document
  - [ ] For range of document
  - [ ] Delta
- [x] Completion
- [x] Definition
- [x] References
- [x] TypeDefinition - markdown file by person name in file path like `families/Snow/Jon.md`
- [x] Hover hints
- [x] DocumentHighlight - means background highlight all references of currently focused person name
- [x] Rename
- [x] Folding
- [x] CodeAction
  - [x] QuickFix for "Unknown family" error
  - [x] QuickFix for "An unobvious name" warning
- [x] Symbol
  - [x] For current document - in editor could be shown in file path toolbar as surname and name of currently focused name like `families/Targaryen.family * Snow * Jon`
  - [x] Fow workspace - helpful to find any person from any place like in vscode by running command `#SnowJon`
- [x] Tree view - helpful to build family tree like
    ```
    Targaryen
      Rhaegar + Stark Lyanna
        Jon
    ```

## Supported languages

- [x] English
- [x] Українська
- [x] Русский


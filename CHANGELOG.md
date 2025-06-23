# Change Log

## [2.1.0] - 2025-06-23

### Changed

- Split build dev server on WebSocket and prod on Stdio

## [2.0.0] - 2025-06-22

### Changed

- Migrate from tree-sitter-familymarkup to familymarkup-parser

## [1.7.0] - 2025-01-12

### Added

- Formating: add child number on `Enter`, remove empty child number on `Enter`
- Completion: label after `=` sign

### Fixed

- Duplicate diagnostic for member with changed surname
- No needed space for next relation while formating
- Bug with completion in unknown member
- Reference name with surname from children of regular relations

## [1.6.0] - 2025-01-07

### Added

- Formating (file, range and on type)

## [1.5.0] - 2025-01-04

### Added

- Setting `warnChildrenWithoutRelations` which enables diagnostic for children without relationships
- Quick fix "Create family relation" for child without relationships

### Removed

- Hint for children in family relation

## [1.4.2] - 2024-12-28

### Fixed

- Document highlights for member reference

## [1.4.1] - 2024-12-27

### Fixed

- duplicates for members with changed surname

## [1.4.0] - 2024-12-27

### Added

- Improved global symbols search - one word will search in families names (and aliases) and members names (and aliases)
- Improved document symbols - now members which changed their surname will be present in that surname document symbols 

### Fixed

- Members with changed surname cosed error because they clone in that surname have Node from Origin member. Now node is refs to first member name reference

## [1.3.0] - 2024-12-27

### Added

- Go to Definition for all surname references
- Template code for changed surname

## [1.2.0] - 2024-12-24

### Added

- Lock in all change places to prevent read/write collisions

### Changed

- Migrate to tree-sitter official package  
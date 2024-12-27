# Change Log

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
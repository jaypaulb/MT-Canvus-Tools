# Icon Set for Custom Menu Designer

This directory contains a minimal icon set for use with the Custom Menu Designer. Icons are organized by category for easy browsing and selection.

## Icon Specifications

- **Format**: PNG
- **Size**: 937x937 pixels (exact dimensions required by Canvus)
- **Color Mode**: RGB or RGBA (with transparency support)
- **File Size**: Optimized (<100KB per icon)

## Directory Structure

```
icons/
├── documents/     # Document and file type icons
├── actions/       # Action icons (create, open, browse, etc.)
├── navigation/    # Navigation icons (home, back, forward, etc.)
└── categories/    # Category and collection icons
```

## Icon Categories

### Documents (`documents/`)
- `document.png` - Generic document icon
- `pdf.png` - PDF document icon
- `image.png` - Image file icon
- `video.png` - Video file icon
- `note.png` - Note/text icon

### Actions (`actions/`)
- `create.png` - Create/Add icon
- `open-folder.png` - Open folder icon
- `browser.png` - Browser/Web icon
- `settings.png` - Settings/Config icon

### Navigation (`navigation/`)
- `home.png` - Home icon
- `back.png` - Back/Previous icon
- `forward.png` - Forward/Next icon

### Categories (`categories/`)
- `folder.png` - Folder icon
- `category.png` - Category/Tag icon
- `collection.png` - Collection icon

## Usage

Icons in this set can be:
1. Selected from the icon picker in the Custom Menu Designer
2. Referenced by relative path in `menu.yml` files
3. Used as templates for custom icons

## Icon Paths

When using these icons in `menu.yml`, use relative paths from the `menu.yml` file location:

```yaml
icon: icons/documents/document.png
icon: icons/actions/create.png
icon: icons/navigation/home.png
```

## Source

Icons are curated from the `canvus-custom-menu` submodule and organized into this minimal set for distribution with Canvus PowerToys.

## Notes

- All icons are verified to be 937x937 pixels (or very close: 933x937 for some icons)
- Icons are optimized for use in Canvus custom menus
- Users can add custom icons to this directory structure
- Icon picker in Custom Menu Designer will browse this directory structure


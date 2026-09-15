# Visual asset guide

The supplied screens are implementation references, not production page images. Recreate their layout, typography, controls, states, and responsive behavior in React and CSS. Do not ship flattened screenshots as interfaces.

## Production assets

Curated content images live in `public/images` and are registered in `src/lib/media/catalog.ts`. Every asset must have descriptive alternative text, a stable identifier, and an explicit category. Pages render them through `next/image` for responsive sizing and optimization.

Current seed assets cover sport, education, travel, and portrait content. They are deliberately small enough for the repository; original high-resolution media will ultimately be served by the media service and object storage.

## Rules

- Keep screenshots and Figma exports outside the application bundle.
- Use semantic HTML and real interactive controls rather than artwork containing text or buttons.
- Avoid external placeholder-image dependencies in committed screens.
- Preserve source licensing and attribution metadata when real catalog ingestion is implemented.
- Provide loading, error, empty, hover, focus, selected, and disabled states for each relevant component.
- Treat visible names, locations, and account data in mockups as test data; do not hard-code real personal information.

# Public-page performance budgets

These budgets apply to production builds measured on a mobile profile with an empty browser cache. CI automation may be added with Lighthouse once the stable deployment preview is available.

| Metric | Budget |
| --- | ---: |
| Largest Contentful Paint | 2.5 seconds |
| Interaction to Next Paint | 200 milliseconds |
| Cumulative Layout Shift | 0.10 |
| Initial JavaScript per public route | 150 KB compressed |
| Initial CSS per public route | 50 KB compressed |
| Hero image | 350 KB |
| Gallery thumbnail | 150 KB |

Use `next/image`, explicit dimensions, responsive `sizes`, local fonts or system stacks, and route-level code splitting. A change exceeding a budget requires a documented exception and follow-up issue.

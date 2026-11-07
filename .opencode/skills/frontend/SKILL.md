---
name: frontend
description: Use when changing the Swiki Angular frontend, UI, components, styles, routes, assets, or API clients. Source lives in D:\develop\angular\swiki\swiki; the Go repository only contains compiled bundles under frontend/.
---

# Swiki Frontend

The Angular source lives at `D:\develop\angular\swiki\swiki` and is registered as the `frontend-src` reference.

Make frontend changes only in the Angular source directory. Never edit the compiled files in `D:\develop\go\swiki\frontend` by hand.

## Development

Run from `D:\develop\angular\swiki\swiki`:

```powershell
npm start
```

## Production build

Build the Angular app:

```powershell
npm run build
```

The output is in `D:\develop\angular\swiki\swiki\dist\swiki\browser`.

Copy the generated files into the Go project's `frontend` directory:

```powershell
Copy-Item -Path 'D:\develop\angular\swiki\swiki\dist\swiki\browser\*' -Destination 'D:\develop\go\swiki\frontend' -Recurse -Force
```

Verify that `frontend/index.html` references the newly generated hashed bundles.

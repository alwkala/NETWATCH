# Third-Party Notices & Licensing

NETWATCH bundles or distributes several third-party software components, fonts, and datasets. This document provides proper attribution, notices, and license texts as required by their respective licenses.

---

## 1. Bundled Typography

### Plus Jakarta Sans
- **Copyright**: Copyright (c) 2020 The Plus Jakarta Sans Project Authors
- **Designer**: Tokotype (Gumpita Rahayu)
- **License**: [SIL Open Font License 1.1](https://openfontlicense.org)
- **Source**: Distributed locally via `@fontsource/plus-jakarta-sans`

```
SIL OPEN FONT LICENSE Version 1.1 - 26 February 2007

PREAMBLE
The goals of the Open Font License (OFL) are to stimulate worldwide development
of collaborative font projects, to support the font creation efforts of academic
and linguistic communities, and to provide a free and open framework in which
fonts may be shared and improved in partnership with others.

PERMISSION & CONDITIONS
Permission is hereby granted, free of charge, to any person obtaining a copy
of the Font Software, to use, study, copy, merge, embed, modify, redistribute,
and sell modified and unmodified copies of the Font Software, subject to the
conditions stated in the SIL Open Font License 1.1.
```

---

### JetBrains Mono
- **Copyright**: Copyright (c) 2020 JetBrains s.r.o.
- **Designer**: Philipp Nurullin, Konstantin Bulenkov
- **License**: [SIL Open Font License 1.1](https://openfontlicense.org)
- **Source**: Distributed locally via `@fontsource/jetbrains-mono`

---

## 2. Bundled Datasets

### IEEE OUI Listing (`internal/oui/ieee-oui.txt`)
- **Publisher**: Institute of Electrical and Electronics Engineers (IEEE) Registration Authority
- **Usage**: Used strictly as an offline, air-gapped hardware vendor identification database without external API queries.
- **License**: Public Standards Data / Free for informational reference.

---

## 3. Core Open-Source Dependencies

### Go Core Dependencies
- **`github.com/ncruces/go-sqlite3`**: csqlite, SQLite driver for Go.  
  *License*: MIT License (Copyright (c) 2022 ncruces). SQLite core is in the Public Domain.
- **`github.com/wailsapp/wails/v2`**: Cross-platform desktop application framework.  
  *License*: MIT License (Copyright (c) 2019-present Lea Anthony).
- **`golang.org/x/sys`**: Go supplemental system libraries.  
  *License*: BSD 3-Clause License (Copyright (c) The Go Authors).

### Frontend Dependencies
- **React & React DOM**: UI framework.  
  *License*: MIT License (Copyright (c) Meta Platforms, Inc.).
- **Lucide Icons (`lucide-react`)**: Clean icons for web development.  
  *License*: ISC License (Copyright (c) Lucide Contributors).
- **Tailwind CSS (`@tailwindcss/vite`, `tailwindcss`)**: Utility-first CSS framework.  
  *License*: MIT License (Copyright (c) Tailwind Labs, Inc.).
- **Vite**: Frontend build tooling.  
  *License*: MIT License (Copyright (c) 2019-present VoidZero Inc. & Vite Contributors).

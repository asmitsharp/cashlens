# cashlens Design System (Pareto Theme)

**Status:** Adopted
**Version:** 1.0

## 1. Philosophy

This document defines the visual identity for the "cashlens" application. The design is a direct implementation of the "Pareto" UI theme.

- **Minimal & Aesthetic:** The UI is clean, spacious, and relies on a limited color palette.
- **Typography-First:** A sophisticated mix of serif and sans-serif fonts differentiates marketing/display text from UI/data text.
- **Professional & Calm:** The color palette is desaturated and calm (greens, grays, blues), building trust and focus.
- **Component-Based:** All UI will be constructed using `shadcn/ui` components, styled to match this specification.

---

## 2. Color Palette

All colors are defined as CSS variables in `app/globals.css` to be used by Tailwind and `shadcn/ui`.

| Role                 | CSS Variable             | HSL Value (Light Theme) | Hex (Approx.) | Description                             |
| :------------------- | :----------------------- | :---------------------- | :------------ | :-------------------------------------- |
| **Background**       | `--background`           | `160 20% 97%`           | `#f7fbfb`     | Main page background (light green-gray) |
| **Foreground**       | `--foreground`           | `160 25% 25%`           | `#2a4a4a`     | Primary text (dark green-teal)          |
| **Card**             | `--card`                 | `160 20% 100%`          | `#ffffff`     | Default card background (white)         |
| **Card (Muted)**     | `--card-secondary`       | `160 30% 94%`           | `#eaf1f1`     | Muted card bg (e.g., active accordion)  |
| **Card (Info)**      | `--card-tertiary`        | `200 25% 95%`           | `#edf4f7`     | Info boxes (e.g., TVL, Price)           |
| **Primary (Accent)** | `--primary`              | `165 77% 64%`           | `#79fced`     | Main CTA, active tabs (bright mint)     |
| **Primary Text**     | `--primary-foreground`   | `160 25% 20%`           | `#223d3d`     | Text on top of the primary accent       |
| **Secondary**        | `--secondary`            | `160 30% 94%`           | `#eaf1f1`     | Muted components, inactive tabs         |
| **Secondary Text**   | `--secondary-foreground` | `160 25% 30%`           | `#325555`     | Text on secondary components            |
| **Muted**            | `--muted`                | `160 20% 90%`           | `#e0e8e8`     | Subtle text, dividers                   |
| **Muted Text**       | `--muted-foreground`     | `160 20% 45%`           | `#617d7d`     | Muted placeholder/info text             |
| **Border**           | `--border`               | `160 20% 88%`           | `#dce5e5`     | Component borders                       |
| **Ring (Focus)**     | `--ring`                 | `160 25% 45%`           | `#587474`     | Focus rings for accessibility           |

---

## 3. Typography

We will use two Google Fonts: **`Inter`** (Sans-Serif) and **`Lora`** (Serif).

| Font           | Variable            | Weight        | Usage                                                                                               |
| :------------- | :------------------ | :------------ | :-------------------------------------------------------------------------------------------------- |
| **Sans-Serif** | `var(--font-sans)`  | 400, 500, 700 | **Default.** All UI text, body copy, sub-headings, nav links, buttons.                              |
| **Serif**      | `var(--font-serif)` | 400, 500      | **Display only.** Main marketing headlines on the landing page (e.g., "Radically transforming..."). |

- **`app/layout.tsx`** must be configured to load these fonts.
- **`tailwind.config.js`** must be updated to include `fontFamily: { sans: ["var(--font-sans)", ...], serif: ["var(--font-serif)", ...] }`.

---

## 4. Border Radius

The theme uses significantly rounded corners.

- **`--radius` (in `globals.css`):** `1.0rem`
- **Tailwind:** `rounded-2xl` and `rounded-3xl` will be used frequently for large cards. Buttons will use `rounded-full` or `rounded-lg`.

---

## 5. Iconography

- **Library:** `lucide-react` (default for `shadcn/ui`).
- **Style:** All icons must be thin, line-art style to match the custom icons in the Pareto screenshots.
- **Custom Icons:** For diagrams (like the "How it Works" flow), custom SVG icons matching the line-art style are required.

---

## 6. Core Component Specifications

### Landing Page (`app/page.tsx`)

- **Navbar:** Minimal. Logo left. "Product", "Features" links (sans-serif, uppercase). "Enter App" button (Primary Accent color).
- **Hero:** Center-aligned. `font-serif` for main headline, `font-sans` for sub-headline.
- **Feature Cards:** Use `shadcn/ui` Card, styled with `rounded-2xl` or `rounded-3xl` and soft shadows.
- **Info Boxes:** Use the `--card-tertiary` (light blue) background.
- **Pill Tabs:** Use `shadcn/ui` Toggle Group, styled to match (active state is `--primary`).
- **Buttons:**
  - **Primary:** Mint accent (`--primary`).
  - **Secondary:** Outlined, dark text, dark border (`--border` or `--foreground`).

### Authentication (`app/(auth)/...`)

- **Provider:** Clerk (as planned).
- **UI:** Use Clerk's pre-built components (`<SignIn />`, `<SignUp />`).
- **Styling:** The components must be styled using Clerk's `appearance` prop to match this design system:
  - `rootBox`: Center on the page.
  - `card`: Set background to `--card` (`#ffffff`), `borderRadius` to `1.0rem`, and add a soft shadow.
  - `formButtonPrimary`: Set background to `--primary` (mint), text to `--primary-foreground`.
  - `formFieldInput`: Set border to `--border`.

### Dashboard Layout (`app/(dashboard)/layout.tsx`)

- **Navbar:** A simple, thin top-bar is preferred over a sidebar to maintain the minimal aesthetic.
  - Logo, "Dashboard", "Review", "Upload".
  - User profile button (Clerk's `<UserButton />`) on the far right.
- **Content:** Main content area will have a `bg-background` (`#f7fbfb`) color.

### Dashboard Page (`app/(dashboard)/page.tsx`)

- **KPI Cards:** Use `shadcn/ui` Card. `rounded-2xl`. White background (`--card`).
- **Charts (Recharts):**
  - **Bar/Line Colors:** Use `--foreground` (`#2a4a4a`) or `--foreground-secondary` (`#587474`).
  * **Positive (Inflow):** Use a clear, positive green (can be a new color, e.g., `hsl(140, 50%, 50%)`).
  * **Negative (Outflow):** Use a clear, negative red/orange (e.g., `hsl(0, 70%, 60%)`).
- **Data Tables:** Use `shadcn/ui` Table. Minimalist, no strong borders.

### Smart Review (`/review`)

- **UI:** A full-page `shadcn/ui` Data Table.
- **Interactive Component:** The "Category" column will use a `shadcn/ui` **Combobox** (autocomplete dropdown) styled to match the theme.

### Profile (`/profile`)

- **UI:** Use Clerk's `<UserProfile />` component.
- **Styling:** Apply the same `appearance` prop as the Sign-In page to make it match the theme (rounded cards, correct button colors).

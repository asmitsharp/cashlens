# CashLens Design System (Pareto Theme)

**Status:** Active - Single Source of Truth
**Version:** 1.0
**Last Updated:** 2025-11-05

## 1. Philosophy

This document is the **single source of truth** for all UI/UX in CashLens. Every frontend component, page, and style must strictly adhere to this specification.

**Design Principles:**
- **Simple:** No unnecessary complexity or decoration
- **Minimal:** Generous whitespace, clean layouts, focused content
- **Spacious:** Comfortable padding and margins throughout
- **Professional:** Enterprise-ready aesthetics for SMB founders
- **Accessible:** WCAG 2.1 AA compliant (4.5:1 contrast minimum)

**Tech Stack:**
- Component Library: `shadcn/ui`
- Styling: Tailwind CSS with CSS variables
- Fonts: Inter (UI), Lora (landing page headlines only)
- Icons: Lucide React

---

## 2. Color Palette

All colors are defined as CSS variables in `cashlens-web/app/globals.css`. This enables consistent theming across all `shadcn/ui` components.

### Core Colors

| Role                 | CSS Variable             | HSL Value           | Usage                                  |
| :------------------- | :----------------------- | :------------------ | :------------------------------------- |
| **Background**       | `--background`           | `0 0% 100%`         | Pure white - page background           |
| **Foreground**       | `--foreground`           | `240 10% 3.9%`      | Near black - primary text              |
| **Card**             | `--card`                 | `0 0% 100%`         | White - card backgrounds               |
| **Card Foreground**  | `--card-foreground`      | `240 10% 3.9%`      | Near black - card text                 |
| **Popover**          | `--popover`              | `0 0% 100%`         | White - dropdown backgrounds           |
| **Popover Foreground** | `--popover-foreground` | `240 10% 3.9%`      | Near black - dropdown text             |
| **Primary**          | `--primary`              | `240 5.9% 10%`      | Near black - primary actions, buttons  |
| **Primary Foreground** | `--primary-foreground` | `0 0% 98%`          | White text on primary buttons          |
| **Secondary**        | `--secondary`            | `240 4.8% 95.9%`    | Light gray - secondary buttons         |
| **Secondary Foreground** | `--secondary-foreground` | `240 5.9% 10%` | Dark text on secondary buttons         |
| **Muted**            | `--muted`                | `240 4.8% 95.9%`    | Light gray - disabled states           |
| **Muted Foreground** | `--muted-foreground`     | `240 3.8% 46.1%`    | Medium gray - secondary text           |
| **Accent**           | `--accent`               | `240 4.8% 95.9%`    | Light gray - hover states              |
| **Accent Foreground** | `--accent-foreground`   | `240 5.9% 10%`      | Dark text on accents                   |
| **Destructive**      | `--destructive`          | `0 84.2% 60.2%`     | Red - errors, delete actions           |
| **Destructive Foreground** | `--destructive-foreground` | `0 0% 98%` | White text on red                      |
| **Border**           | `--border`               | `240 5.9% 90%`      | Light gray - component borders         |
| **Input**            | `--input`                | `240 5.9% 90%`      | Light gray - input borders             |
| **Ring**             | `--ring`                 | `240 5.9% 10%`      | Near black - focus rings               |

### Custom Semantic Colors

| Role             | CSS Variable      | HSL Value         | Usage                          |
| :--------------- | :---------------- | :---------------- | :----------------------------- |
| **Success**      | `--success`       | `142 76% 36%`     | Green - positive cash flow     |
| **Success FG**   | `--success-foreground` | `0 0% 98%` | White text on green            |
| **Warning**      | `--warning`       | `38 92% 50%`      | Amber - warnings, pending      |
| **Warning FG**   | `--warning-foreground` | `0 0% 98%`  | White text on amber            |
| **Chart Green**  | `--chart-green`   | `142 76% 36%`     | Positive values in charts      |
| **Chart Red**    | `--chart-red`     | `0 84.2% 60.2%`   | Negative values in charts      |
| **Chart Blue**   | `--chart-blue`    | `221 83% 53%`     | Neutral data in charts         |
| **Chart Amber**  | `--chart-amber`   | `38 92% 50%`      | Highlights in charts           |
| **Chart Purple** | `--chart-purple`  | `262 83% 58%`     | Secondary data in charts       |

---

## 3. Typography

### Font Families

**Inter (Sans-Serif)** - Primary UI font
- Usage: All dashboard UI, buttons, labels, body text, navigation, data tables
- Weights: 400 (Regular), 500 (Medium), 600 (Semi-bold), 700 (Bold)
- Variable: `var(--font-sans)` or `font-sans` Tailwind class

**Lora (Serif)** - Display font
- Usage: **Landing page headlines ONLY** (not in dashboard)
- Weights: 600 (Semi-bold), 700 (Bold)
- Variable: `var(--font-serif)` or `font-serif` Tailwind class

### Font Scale (Tailwind Classes)

```css
text-xs:    0.75rem  (12px)  /* Small labels, captions */
text-sm:    0.875rem (14px)  /* Secondary text, table cells */
text-base:  1rem     (16px)  /* Body text, default */
text-lg:    1.125rem (18px)  /* Large body text */
text-xl:    1.25rem  (20px)  /* Section subheadings */
text-2xl:   1.5rem   (24px)  /* Card titles */
text-3xl:   1.875rem (30px)  /* Page titles */
text-4xl:   2.25rem  (36px)  /* Dashboard headlines */
```

### Line Heights

```css
leading-tight:   1.25   /* Headlines */
leading-snug:    1.375  /* Subheadings */
leading-normal:  1.5    /* Body text */
leading-relaxed: 1.625  /* Long-form content */
```

### Font Weights

```css
font-normal:    400  /* Body text */
font-medium:    500  /* Labels, emphasized text */
font-semibold:  600  /* Button text, card titles */
font-bold:      700  /* Page headings */
```

---

## 4. Spacing Scale

Use Tailwind's default spacing (1 unit = 0.25rem = 4px):

```css
p-2:  0.5rem   (8px)   /* Tight spacing */
p-4:  1rem     (16px)  /* Standard padding */
p-6:  1.5rem   (24px)  /* Card padding */
p-8:  2rem     (32px)  /* Page padding */
p-12: 3rem     (48px)  /* Section padding */

gap-2:  0.5rem   /* Tight gaps */
gap-4:  1rem     /* Standard gaps */
gap-6:  1.5rem   /* Card gaps */
gap-8:  2rem     /* Section gaps */
```

**Layout Guidelines:**
- Page margins: `mx-auto max-w-7xl px-4 sm:px-6 lg:px-8`
- Card padding: `p-6` (24px)
- Section spacing: `space-y-8` (32px vertical)
- Form spacing: `space-y-4` (16px between fields)
- Button padding: `px-4 py-2` (16px h, 8px v)

---

## 5. Border Radius

**Base radius:** `--radius: 1.0rem` (16px) in `globals.css`

```css
rounded-lg:    0.5rem   (8px)   /* Buttons, inputs */
rounded-xl:    0.75rem  (12px)  /* Medium elements */
rounded-2xl:   1rem     (16px)  /* Cards, modals - DEFAULT */
rounded-3xl:   1.5rem   (24px)  /* Hero sections */
rounded-full:  9999px           /* Circular (avatars, pills) */
```

**Usage:**
- Cards: `rounded-2xl` (16px)
- Buttons: `rounded-lg` (8px)
- Inputs: `rounded-lg` (8px)
- Modals: `rounded-2xl` (16px)
- Badges: `rounded-md` (6px)

---

## 6. Shadows

```css
shadow-sm:   Subtle, minimal elevation
shadow:      Default cards
shadow-md:   Elevated cards
shadow-lg:   Modals, dialogs
shadow-xl:   Popovers, dropdowns
shadow-2xl:  Hero elements
```

**Usage:**
- Cards: `shadow` or `shadow-md`
- Hover: Increase shadow (`hover:shadow-lg`)
- Modals: `shadow-2xl`

---

## 7. Iconography

- **Library:** `lucide-react` (default for shadcn/ui)
- **Style:** Line-art, minimal, consistent stroke width

**Icon Sizes:**
```tsx
w-4 h-4   /* 16px - Inline text */
w-5 h-5   /* 20px - Buttons (default) */
w-6 h-6   /* 24px - Headers */
w-8 h-8   /* 32px - Hero */
```

---

## 8. Component Specifications

### Buttons

**Primary Button:**
```tsx
<Button className="bg-primary text-primary-foreground hover:bg-primary/90 rounded-lg">
  Upload CSV
</Button>
```

**Secondary Button:**
```tsx
<Button variant="secondary" className="rounded-lg">
  Cancel
</Button>
```

**Sizes:** Small: `px-3 py-1.5 text-sm`, Default: `px-4 py-2`, Large: `px-6 py-3 text-lg`

### Cards

```tsx
<Card className="rounded-2xl p-6 shadow">
  <CardHeader>
    <CardTitle className="text-2xl font-semibold">Title</CardTitle>
  </CardHeader>
  <CardContent>{/* Content */}</CardContent>
</Card>
```

### Inputs

```tsx
<Input
  type="text"
  placeholder="Enter email"
  className="rounded-lg border-input"
/>
```

---

## 9. Clerk Authentication Theming

**Appearance configuration for sign-in/sign-up pages:**

```tsx
const clerkAppearance = {
  elements: {
    formButtonPrimary:
      'bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-semibold',
    card: 'rounded-2xl shadow-md',
    headerTitle: 'text-2xl font-bold text-foreground',
    headerSubtitle: 'text-muted-foreground',
    formFieldLabel: 'text-sm font-medium text-foreground',
    formFieldInput: 'rounded-lg border-input focus:ring-2 focus:ring-ring',
    footerActionLink: 'text-primary hover:text-primary/90',
  },
  variables: {
    colorPrimary: 'hsl(240, 5.9%, 10%)',
    colorBackground: 'hsl(0, 0%, 100%)',
    borderRadius: '0.5rem',
  },
}
```

---

## 10. Implementation Checklist

When creating UI components, ensure:

- [ ] Uses Inter font (`font-sans`)
- [ ] Colors use CSS variables (not hardcoded)
- [ ] Cards use `rounded-2xl`, buttons use `rounded-lg`
- [ ] Proper spacing (p-6 for cards, p-4 for standard)
- [ ] Hover states with transitions
- [ ] Focus rings for accessibility
- [ ] WCAG 2.1 AA contrast compliance
- [ ] Keyboard accessible
- [ ] Mobile responsive

---

**This is the single source of truth for all CashLens UI. All components must follow these specifications.**

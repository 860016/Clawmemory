<template>
  <router-view />
</template>

<style>
/* ===== ClawMemory Theme System ===== */
/* Dark theme (default) */
:root,
[data-theme="dark"] {
  --cm-primary: #10B981;
  --cm-primary-light: #34d399;
  --cm-primary-dark: #059669;
  --cm-primary-rgb: 16, 185, 129;

  --cm-bg: #0d1117;
  --cm-bg-secondary: #161b22;
  --cm-bg-tertiary: #1c2129;
  --cm-text: #e6edf3;
  --cm-text-secondary: #c9d1d9;
  --cm-text-muted: #7d8590;
  --cm-text-placeholder: #484f58;
  --cm-border: #21262d;
  --cm-border-light: #30363d;

  --cm-shadow: 0 2px 12px rgba(0, 0, 0, 0.3);
  --cm-shadow-lg: 0 8px 32px rgba(0, 0, 0, 0.4);
  --cm-radius: 12px;
  --cm-radius-sm: 8px;
  --cm-radius-lg: 16px;

  /* Element Plus overrides */
  --el-bg-color: #0d1117;
  --el-bg-color-overlay: #161b22;
  --el-bg-color-page: #0d1117;
  --el-text-color-primary: #e6edf3;
  --el-text-color-regular: #c9d1d9;
  --el-text-color-secondary: #7d8590;
  --el-text-color-placeholder: #484f58;
  --el-border-color: #21262d;
  --el-border-color-light: #21262d;
  --el-border-color-lighter: #161b22;
  --el-fill-color: #161b22;
  --el-fill-color-light: #21262d;
  --el-fill-color-lighter: #161b22;
  --el-fill-color-blank: #0d1117;
  --el-color-primary: #10B981;
  --el-color-primary-light-3: #34d399;
  --el-color-primary-light-5: #6ee7b7;
  --el-color-primary-light-7: #a7f3d0;
  --el-color-primary-light-9: #d1fae5;
  --el-color-primary-dark-2: #059669;
  --el-mask-color: rgba(1, 4, 9, 0.7);
}

/* Light theme */
[data-theme="light"] {
  --cm-primary: #10B981;
  --cm-primary-light: #34d399;
  --cm-primary-dark: #059669;
  --cm-primary-rgb: 16, 185, 129;

  --cm-bg: #f8fafc;
  --cm-bg-secondary: #ffffff;
  --cm-bg-tertiary: #f1f5f9;
  --cm-text: #1e293b;
  --cm-text-secondary: #475569;
  --cm-text-muted: #94a3b8;
  --cm-text-placeholder: #cbd5e1;
  --cm-border: #e2e8f0;
  --cm-border-light: #cbd5e1;

  --cm-shadow: 0 2px 12px rgba(0, 0, 0, 0.06);
  --cm-shadow-lg: 0 8px 32px rgba(0, 0, 0, 0.1);
  --cm-radius: 12px;
  --cm-radius-sm: 8px;
  --cm-radius-lg: 16px;

  /* Element Plus overrides */
  --el-bg-color: #ffffff;
  --el-bg-color-overlay: #f8fafc;
  --el-bg-color-page: #f8fafc;
  --el-text-color-primary: #1e293b;
  --el-text-color-regular: #475569;
  --el-text-color-secondary: #94a3b8;
  --el-text-color-placeholder: #cbd5e1;
  --el-border-color: #e2e8f0;
  --el-border-color-light: #e2e8f0;
  --el-border-color-lighter: #f1f5f9;
  --el-fill-color: #f8fafc;
  --el-fill-color-light: #f1f5f9;
  --el-fill-color-lighter: #f8fafc;
  --el-fill-color-blank: #ffffff;
  --el-color-primary: #10B981;
  --el-color-primary-light-3: #34d399;
  --el-color-primary-light-5: #6ee7b7;
  --el-color-primary-light-7: #a7f3d0;
  --el-color-primary-light-9: #d1fae5;
  --el-color-primary-dark-2: #059669;
  --el-mask-color: rgba(0, 0, 0, 0.5);
}

body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: var(--cm-bg);
  color: var(--cm-text);
  transition: background 0.3s ease, color 0.3s ease;
}

/* ===== Element Plus Dark/Light Theme Overrides ===== */

/* Dialog */
.el-dialog {
  --el-dialog-bg-color: var(--cm-bg-secondary) !important;
  --el-dialog-title-font-size: 16px;
  border: 1px solid var(--cm-border) !important;
  border-radius: var(--cm-radius-lg) !important;
}
.el-dialog__header {
  border-bottom: 1px solid var(--cm-border);
  padding-bottom: 12px;
}
.el-dialog__title {
  color: var(--cm-text) !important;
}

/* Input */
.el-input__wrapper {
  background-color: var(--cm-bg) !important;
  box-shadow: 0 0 0 1px var(--cm-border) inset !important;
  transition: box-shadow 0.2s ease;
}
.el-input__wrapper:hover {
  box-shadow: 0 0 0 1px var(--cm-border-light) inset !important;
}
.el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px var(--cm-primary) inset !important;
}
.el-input__inner {
  color: var(--cm-text) !important;
}
.el-input__inner::placeholder {
  color: var(--cm-text-placeholder) !important;
}

/* Textarea */
.el-textarea__inner {
  background-color: var(--cm-bg) !important;
  color: var(--cm-text) !important;
  border-color: var(--cm-border) !important;
}
.el-textarea__inner:focus {
  border-color: var(--cm-primary) !important;
}

/* Select */
.el-select .el-input__wrapper {
  background-color: var(--cm-bg) !important;
}
.el-select-dropdown {
  background-color: var(--cm-bg-secondary) !important;
  border-color: var(--cm-border) !important;
}

/* Button primary */
.el-button--primary {
  --el-button-bg-color: #10B981;
  --el-button-border-color: #10B981;
  --el-button-hover-bg-color: #34d399;
  --el-button-hover-border-color: #34d399;
  --el-button-active-bg-color: #059669;
  --el-button-active-border-color: #059669;
}

[data-theme="dark"] .el-button--primary {
  --el-button-text-color: #0d1117;
}

[data-theme="light"] .el-button--primary {
  --el-button-text-color: #ffffff;
}

/* Radio button */
.el-radio-button__inner {
  background: var(--cm-bg-secondary) !important;
  border-color: var(--cm-border) !important;
  color: var(--cm-text-muted) !important;
}
.el-radio-button__original-radio:checked + .el-radio-button__inner {
  background: rgba(var(--cm-primary-rgb), 0.15) !important;
  border-color: var(--cm-primary) !important;
  color: var(--cm-primary) !important;
  box-shadow: -1px 0 0 0 var(--cm-primary) !important;
}

/* Tabs */
.el-tabs__item {
  color: var(--cm-text-muted) !important;
}
.el-tabs__item.is-active {
  color: var(--cm-primary) !important;
}
.el-tabs__active-bar {
  background-color: var(--cm-primary) !important;
}
.el-tabs__nav-wrap::after {
  background-color: var(--cm-border) !important;
}

/* Pagination */
.el-pagination {
  --el-pagination-bg-color: var(--cm-bg-secondary);
  --el-pagination-text-color: var(--cm-text-muted);
  --el-pagination-button-bg-color: var(--cm-bg-secondary);
}
.el-pager li {
  background: var(--cm-bg-secondary) !important;
  color: var(--cm-text-muted) !important;
}
.el-pager li.is-active {
  background: rgba(var(--cm-primary-rgb), 0.15) !important;
  color: var(--cm-primary) !important;
}

/* Badge */
.el-badge__content {
  background-color: var(--cm-border) !important;
  color: var(--cm-text-muted) !important;
  border: none !important;
}

/* Slider */
.el-slider__bar {
  background-color: var(--cm-primary) !important;
}
.el-slider__button {
  border-color: var(--cm-primary) !important;
}

/* Upload */
.el-upload-dragger {
  background-color: var(--cm-bg-secondary) !important;
  border-color: var(--cm-border) !important;
}

/* MessageBox */
.el-message-box {
  --el-messagebox-title-color: var(--cm-text);
  --el-messagebox-content-color: var(--cm-text-secondary);
  background-color: var(--cm-bg-secondary) !important;
  border: 1px solid var(--cm-border) !important;
}

/* Form item label */
.el-form-item__label {
  color: var(--cm-text-secondary) !important;
}

/* Switch */
.el-switch.is-checked .el-switch__core {
  background-color: var(--cm-primary) !important;
  border-color: var(--cm-primary) !important;
}

/* Tag */
.el-tag {
  border-radius: 6px !important;
}

/* Card */
.el-card {
  border-radius: var(--cm-radius) !important;
  border-color: var(--cm-border) !important;
  box-shadow: var(--cm-shadow) !important;
  background: var(--cm-bg-secondary) !important;
}

/* Table */
.el-table {
  --el-table-bg-color: var(--cm-bg-secondary);
  --el-table-tr-bg-color: var(--cm-bg-secondary);
  --el-table-header-bg-color: var(--cm-bg-tertiary);
  --el-table-row-hover-bg-color: var(--cm-bg-tertiary);
  --el-table-border-color: var(--cm-border);
  --el-table-text-color: var(--cm-text-secondary);
  --el-table-header-text-color: var(--cm-text);
}
.el-table th.el-table__cell {
  background-color: var(--cm-bg-tertiary) !important;
  color: var(--cm-text) !important;
  font-weight: 600;
}

/* Dropdown */
.el-dropdown-menu {
  background-color: var(--cm-bg-secondary) !important;
  border-color: var(--cm-border) !important;
}
.el-dropdown-menu__item {
  color: var(--cm-text-secondary) !important;
}
.el-dropdown-menu__item:hover {
  background-color: var(--cm-bg-tertiary) !important;
  color: var(--cm-text) !important;
}

/* Alert */
.el-alert {
  border-radius: var(--cm-radius-sm) !important;
}

/* Scrollbar - Dark */
[data-theme="dark"] ::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
[data-theme="dark"] ::-webkit-scrollbar-track {
  background: #0d1117;
}
[data-theme="dark"] ::-webkit-scrollbar-thumb {
  background: #30363d;
  border-radius: 4px;
}
[data-theme="dark"] ::-webkit-scrollbar-thumb:hover {
  background: #484f58;
}

/* Scrollbar - Light */
[data-theme="light"] ::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}
[data-theme="light"] ::-webkit-scrollbar-track {
  background: #f1f5f9;
}
[data-theme="light"] ::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 4px;
}
[data-theme="light"] ::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* Transitions for theme switching */
*,
*::before,
*::after {
  transition-property: background-color, border-color, color, box-shadow;
  transition-duration: 0.2s;
  transition-timing-function: ease;
}

/* Exclude transition from elements that need instant response */
input,
textarea,
[contenteditable] {
  transition-property: background-color, border-color, box-shadow;
}
</style>

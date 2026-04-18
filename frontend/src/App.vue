<template>
  <router-view />
</template>

<style>
/* ClawMemory Dark Theme - Global Override */
:root {
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
  --el-color-primary: #00d4aa;
  --el-color-primary-light-3: #33deb8;
  --el-color-primary-light-5: #4de6c4;
  --el-color-primary-light-7: #99edda;
  --el-color-primary-light-9: #ccf5ec;
  --el-color-primary-dark-2: #00aa88;
  --el-mask-color: rgba(1, 4, 9, 0.7);
}

body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: #0d1117;
  color: #e6edf3;
}

/* Dialog dark */
.el-dialog {
  --el-dialog-bg-color: #161b22 !important;
  --el-dialog-title-font-size: 16px;
  border: 1px solid #21262d !important;
}

.el-dialog__header {
  border-bottom: 1px solid #21262d;
  padding-bottom: 12px;
}

.el-dialog__title {
  color: #e6edf3 !important;
}

/* Input dark */
.el-input__wrapper {
  background-color: #0d1117 !important;
  box-shadow: 0 0 0 1px #21262d inset !important;
}

.el-input__wrapper:hover {
  box-shadow: 0 0 0 1px #30363d inset !important;
}

.el-input__wrapper.is-focus {
  box-shadow: 0 0 0 1px #00d4aa inset !important;
}

.el-input__inner {
  color: #e6edf3 !important;
}

.el-input__inner::placeholder {
  color: #484f58 !important;
}

/* Textarea */
.el-textarea__inner {
  background-color: #0d1117 !important;
  color: #e6edf3 !important;
  border-color: #21262d !important;
}

.el-textarea__inner:focus {
  border-color: #00d4aa !important;
}

/* Select */
.el-select .el-input__wrapper {
  background-color: #0d1117 !important;
}

/* Button primary */
.el-button--primary {
  --el-button-bg-color: #00d4aa;
  --el-button-border-color: #00d4aa;
  --el-button-hover-bg-color: #00e6b8;
  --el-button-hover-border-color: #00e6b8;
  --el-button-active-bg-color: #00aa88;
  --el-button-active-border-color: #00aa88;
  --el-button-text-color: #0d1117;
}

/* Radio button */
.el-radio-button__inner {
  background: #161b22 !important;
  border-color: #21262d !important;
  color: #7d8590 !important;
}

.el-radio-button__original-radio:checked + .el-radio-button__inner {
  background: rgba(0, 212, 170, 0.15) !important;
  border-color: #00d4aa !important;
  color: #00d4aa !important;
  box-shadow: -1px 0 0 0 #00d4aa !important;
}

/* Tabs */
.el-tabs__item {
  color: #7d8590 !important;
}

.el-tabs__item.is-active {
  color: #00d4aa !important;
}

.el-tabs__active-bar {
  background-color: #00d4aa !important;
}

.el-tabs__nav-wrap::after {
  background-color: #21262d !important;
}

/* Pagination */
.el-pagination {
  --el-pagination-bg-color: #161b22;
  --el-pagination-text-color: #7d8590;
  --el-pagination-button-bg-color: #161b22;
}

.el-pager li {
  background: #161b22 !important;
  color: #7d8590 !important;
}

.el-pager li.is-active {
  background: rgba(0, 212, 170, 0.15) !important;
  color: #00d4aa !important;
}

/* Badge */
.el-badge__content {
  background-color: #21262d !important;
  color: #7d8590 !important;
  border: none !important;
}

/* Slider */
.el-slider__bar {
  background-color: #00d4aa !important;
}

.el-slider__button {
  border-color: #00d4aa !important;
}

/* Upload */
.el-upload-dragger {
  background-color: #161b22 !important;
  border-color: #21262d !important;
}

/* MessageBox */
.el-message-box {
  --el-messagebox-title-color: #e6edf3;
  --el-messagebox-content-color: #c9d1d9;
  background-color: #161b22 !important;
  border: 1px solid #21262d !important;
}

/* Form item label */
.el-form-item__label {
  color: #c9d1d9 !important;
}

/* Scrollbar */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #0d1117;
}

::-webkit-scrollbar-thumb {
  background: #30363d;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #484f58;
}
</style>

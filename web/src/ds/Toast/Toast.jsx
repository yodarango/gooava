import React from "react";
import { createPortal } from "react-dom";
import { If } from "@ds";

import "./Toast.css";

export const Toast = (props) => {
  const { icon, onClose, children, title, type, style, open = false } = props;
  let cardClass;

  switch (type) {
    case "success":
      cardClass = "toast-success";
      break;
    case "danger":
      cardClass = "toast-danger";
      break;
    case "warning":
      cardClass = "toast-warning";
      break;
    case "info":
      cardClass = "toast-info";
      break;
    case "default":
      cardClass = "toast-default";
      break;
    default:
      cardClass = "toast-default";
      break;
  }

  const closeClass = !!onClose ? "py-3 ps-3 pe-1" : "p-3";

  if (!open) return null;

  const { zIndex = 22 } = props;

  const toastContent = (
    <div
      className={`toast-04hl ${closeClass} ${cardClass} d-flex align-items-center justify-content-start gap-3 color-beta`}
      style={{ zIndex, ...props.style }}
    >
      {icon && (
        <div className='color-beta flex-shrink-0'>
          <ion-icon name={icon} />
        </div>
      )}
      <div className='w-100'>
        <div className='d-flex align-items-start justify-content-end column-gap-4'>
          <div className='w-100'>
            <h5>{title}</h5>
            <p>{children}</p>
          </div>
          <If condition={!!onClose}>
            <button
              className='btn-base flex-shrink-0 color-beta'
              onClick={onClose}
            >
              <ion-icon name='close' />
            </button>
          </If>
        </div>
      </div>
    </div>
  );

  return createPortal(toastContent, document.body);
};

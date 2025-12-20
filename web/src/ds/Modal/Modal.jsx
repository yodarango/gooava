import { createPortal } from "react-dom";
import React from "react";

import "./Modal.css";

export const Modal = ({
  closeOnBackdropClick = true,
  contentContainerStyle = {},
  showCloseButton = true,
  showWaves = true,
  height = "auto",
  zIndex = 10,
  children,
  onClose,
  title,
  open,
  ...rest
}) => {
  if (!open) return null;

  const handleBackdropClick = (e) => {
    if (closeOnBackdropClick && onClose) {
      e.stopPropagation();
      onClose(e);
    } else {
      e.stopPropagation();
    }
  };

  // handle the click on the content propagation unless it is the close button or the closeOnBackdropClick is true
  const handleContentClick = (e) => {
    if (e.target.classList.contains("shrood-modal-0elj__close")) {
      onClose();
    } else {
      e.stopPropagation();
    }
  };

  const modalContent = (
    <div className='shrood-modal-0elj' style={{ zIndex: zIndex + 1 }} {...rest}>
      <div
        className='shrood-modal-0elj__content p-6'
        onClick={handleContentClick}
        style={{
          zIndex: zIndex + 1,
          height,
        }}
      >
        {showCloseButton && (
          <button
            className='shrood-modal-0elj__close color-alpha bg-nu'
            onClick={onClose}
          >
            <ion-icon name='close-outline' />
          </button>
        )}
        <h4 className='mb-2 text-center mb-6 px-4'>{title}</h4>
        <div
          className='shrood-modal-0elj-content__content'
          style={{ ...contentContainerStyle }}
        >
          {children}
        </div>
      </div>
      <div
        className='shrood-modal-0elj__backdrop'
        onClick={handleBackdropClick}
        style={{ zIndex }}
      ></div>
    </div>
  );

  // Create portal to render modal at the end of document body
  return createPortal(modalContent, document.body);
};

import React, { useRef, useEffect } from "react";
import "./TextArea.css";

export const TextArea = ({
  primary,
  secondary,
  danger,
  warning,
  success,
  className = "",
  shineOnFocus = true,
  onPressEnter,
  minRows = 1,
  maxRows,
  maxCharacters,
  showCharacterCount = false,
  ...rest
}) => {
  const textAreaRef = useRef(null);

  const adjustHeight = () => {
    const textArea = textAreaRef.current;
    if (!textArea) return;

    // Reset height to auto to get the correct scrollHeight
    textArea.style.height = "auto";

    // Calculate the number of rows based on content
    const lineHeight = parseInt(window.getComputedStyle(textArea).lineHeight);
    const padding = parseInt(window.getComputedStyle(textArea).paddingTop) * 2;
    const contentHeight = textArea.scrollHeight - padding;
    const rows = Math.ceil(contentHeight / lineHeight);

    // Apply min and max row constraints
    const constrainedRows = Math.max(
      minRows,
      maxRows ? Math.min(rows, maxRows) : rows
    );
    const newHeight = constrainedRows * lineHeight + padding;

    textArea.style.height = `${newHeight}px`;
  };

  useEffect(() => {
    adjustHeight();
  }, [rest.value, rest.defaultValue]);

  const handleInput = (e) => {
    // Enforce character limit if maxCharacters is set
    if (maxCharacters && e.target.value.length > maxCharacters) {
      e.target.value = e.target.value.slice(0, maxCharacters);
    }

    adjustHeight();
    if (rest.onInput) {
      rest.onInput(e);
    }
  };

  const handleKeyDown = (e) => {
    // check if they are pressing shift plus enter in which case onPressEnter should be called
    if (e.shiftKey && e.key === "Enter") {
      e.preventDefault(); // Prevent default new line behavior
      if (onPressEnter) {
        onPressEnter(e);
      }
      return;
    }
  };

  const handleChange = (e) => {
    // Enforce character limit if maxCharacters is set
    if (maxCharacters && e.target.value.length > maxCharacters) {
      e.target.value = e.target.value.slice(0, maxCharacters);
    }

    adjustHeight();
    if (rest.onChange) {
      rest.onChange(e);
    }
  };

  // Build className based on color props
  let finalClassName = `custom-textarea ${className}`.trim();

  const shineOnFocusClass = shineOnFocus ? "shine-on-focus" : "";
  if (shineOnFocusClass) finalClassName += ` ${shineOnFocusClass}`;

  if (secondary) finalClassName += " secondary";
  else if (danger) finalClassName += " danger";
  else if (warning) finalClassName += " warning";
  else if (success) finalClassName += " success";
  else finalClassName += " primary";

  // Get current character count
  const currentValue = rest.value || rest.defaultValue || "";
  const currentLength = currentValue.length;
  const remainingCharacters = maxCharacters
    ? maxCharacters - currentLength
    : null;

  return (
    <div className='textarea-container'>
      <textarea
        ref={textAreaRef}
        {...rest}
        className={finalClassName}
        onInput={handleInput}
        onChange={handleChange}
        onKeyDown={handleKeyDown}
        rows={minRows}
        maxLength={maxCharacters}
      />
      {(showCharacterCount || maxCharacters) && (
        <div className='character-count-overlay'>
          <span
            className={
              remainingCharacters !== null && remainingCharacters < 0
                ? "color-dnager"
                : remainingCharacters !== null && remainingCharacters < 20
                ? "color-warning"
                : ""
            }
          >
            {maxCharacters ? `${remainingCharacters}` : `${currentLength}`}
          </span>
        </div>
      )}
    </div>
  );
};

export default TextArea;

import React from "react";

// styles
import "./Switch.css";

export const Switch = ({
  disabled = false,
  checked = false,
  onChange,
  primary,
  secondary,
  danger,
  warning,
  success,
  color,
  className = "",
}) => {
  const handleToggle = () => {
    if (!disabled && onChange) {
      onChange(!checked);
    }
  };

  // Build className for color variants and thumb
  let switchClassName = `custom-switch ${checked ? "checked" : ""} ${
    disabled ? "disabled" : ""
  } ${className}`;

  let thumbClassName = `custom-switch-thumb`;

  if (primary) {
    switchClassName += " switch-primary";
    thumbClassName += " switch-primary-thumb";
  } else if (secondary) {
    switchClassName += " switch-secondary";
    thumbClassName += " switch-secondary-thumb";
  } else if (danger) {
    switchClassName += " switch-danger";
    thumbClassName += " switch-danger-thumb";
  } else if (warning) {
    switchClassName += " switch-warning";
    thumbClassName += " switch-warning-thumb";
  } else if (success) {
    switchClassName += " switch-success";
    thumbClassName += " switch-success-thumb";
  }

  // Handle custom color prop
  const customStyle = color ? { "--switch-custom-color": color } : {};

  return (
    <div
      className={switchClassName}
      onClick={handleToggle}
      role='switch'
      aria-checked={checked}
      aria-disabled={disabled}
      style={customStyle}
    >
      <div className='custom-switch-track'></div>
      <div className={thumbClassName}></div>
      <input
        type='checkbox'
        className='custom-switch-input'
        checked={checked}
        onChange={() => {}}
        disabled={disabled}
      />
    </div>
  );
};

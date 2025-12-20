import React from "react";
import "./Input.css";

export const Input = ({ danger, className, shineOnFocus = true, ...rest }) => {
  const shineOnFocusClass = shineOnFocus ? "shine-on-focus" : "";
  const dangerClass = danger ? "danger" : "";

  return (
    <input
      {...rest}
      className={`custom-input ${dangerClass} ${shineOnFocusClass} ${className}`.trim()}
    />
  );
};

export default Input;

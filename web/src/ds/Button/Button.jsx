import { Loading, IfElse } from "@ds";
import React from "react";
import "./Button.css";

export const Button = (props) => {
  let {
    children,
    primary,
    secondary,
    danger,
    warning,
    success,
    className = "",
    isLoading,
    ...restOfProps
  } = props;

  className = className + " button";

  if (primary) className += " button-primary";
  else if (secondary) className += " button-secondary";
  else if (danger) className += " button-danger";
  else if (warning) className += " button-warning";
  else if (success) className += " button-success";
  else className += " button-default";

  return (
    <button className={className} {...restOfProps}>
      <IfElse condition={isLoading}>
        <Loading size={30} />
        <> {children}</>
      </IfElse>
    </button>
  );
};

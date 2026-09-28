import { SignupForm } from "../components/SignupForm/SignupForm";
import { LoginForm } from "../components/LoginForm/LoginForm";
import React, { useState } from "react";
import { Button, IfElse } from "@ds";
import "./Layout.css";

export const Layout = () => {
  const [isLogin, setIsLogin] = useState(true);

  return (
    <div className='auth-layout-00th'>
      <div className='auth-layout__container-sdn6 bg-gamma rounded-3 p-4'>
        <h3 className='text-center mb-4'>
          {isLogin ? "Welcome Back" : "Create Account"}
        </h3>

        <div className='auth-layout__form mb-4'>
          <IfElse condition={isLogin}>
            <LoginForm />
            <SignupForm />
          </IfElse>
        </div>

        <div className='auth-layout__toggle'>
          <Button
            secondary
            onClick={() => setIsLogin(!isLogin)}
            className='auth-layout__toggle-button w-100'
          >
            {isLogin ? "Sign Up Instead" : "Sign In Instead"}
          </Button>
        </div>
      </div>
    </div>
  );
};

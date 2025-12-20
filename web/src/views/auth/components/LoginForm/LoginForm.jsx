import { useAppContext } from "../../../context/appContextProvider";
import { ForgotPassword } from "../ForogotPassword/ForogotPassword";
import React, { useState, useEffect } from "react";
import { API_POST_LOGIN } from "@constants";
import { Input, Button, Modal } from "@ds";
import { usePost } from "@utils";

// styles
import "./LoginForm.css";

export const LoginForm = () => {
  const { showToast, setupAuth } = useAppContext();

  const [showPassword, setShowPassword] = useState(false);

  const [formData, setFormData] = useState({
    username: "",
    password: "",
  });

  const { post, loading, error } = usePost({
    url: API_POST_LOGIN,
    callback: (data) => {
      if (!data) return;

      localStorage.setItem("auth", data.AuthToken.replace("Bearer ", ""));

      setupAuth();

      showToast({
        type: "success",
        message: "Login successful! Welcome back.",
      });
    },
  });

  const handleInputChange = (field) => (e) => {
    setFormData((prev) => ({
      ...prev,
      [field]: e.target.value,
    }));
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    // Basic validation
    if (!formData.username.trim() || !formData.password.trim()) {
      showToast({
        type: "danger",
        message: "Please fill in all fields",
      });
      return;
    }

    post(formData);
  };

  useEffect(() => {
    if (error) {
      showToast({
        type: "danger",
        message: error,
      });
    }
  }, [error]);

  return (
    <form onSubmit={handleSubmit} className='login-form-38jr'>
      <div className='login-form__field'>
        <label htmlFor='username' className='login-form__label'>
          Username
        </label>
        <Input
          placeholder='Enter your username'
          onChange={handleInputChange("username")}
          className='mb-4'
          value={formData.username}
          id='username'
          type='text'
          required
        />
      </div>

      <div className='login-form__field'>
        <label htmlFor='password' className='login-form__label'>
          Password
        </label>
        <div className='position-relative'>
          <Input
            onChange={handleInputChange("password")}
            placeholder='Enter your password'
            value={formData.password}
            className='mb-4'
            type={showPassword ? "text" : "password"}
            id='password'
            required
          />
          <button
            type='button'
            className='password-toggle-btn position-absolute p-1'
            onClick={() => setShowPassword(!showPassword)}
          >
            <ion-icon
              name={showPassword ? "eye-off-outline" : "eye-outline"}
            ></ion-icon>
          </button>
        </div>
      </div>

      <Button type='submit' primary isLoading={loading} className='w-100 mb-1'>
        Sign In
      </Button>
      <ForgotPassword />
    </form>
  );
};

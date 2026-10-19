import { useAppContext } from "../../../context/appContextProvider";
import { ForgotPassword } from "../ForogotPassword/ForogotPassword";
import React, { useState, useEffect } from "react";
import { API_POST_LOGIN, API_POST_VERIFY_TWO_FA } from "@constants";
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

  // 2FA: when the account has it on, /login answers TwoFARequired instead of
  // a token — the user confirms the emailed 6-digit code to finish signing in
  const [twoFAPending, setTwoFAPending] = useState(false);
  const [twoFACode, setTwoFACode] = useState("");

  const { post, loading, error } = usePost({
    url: API_POST_LOGIN,
    callback: (data) => {
      if (!data) return;

      // 2FA accounts get a code by email instead of an immediate token
      if (data.TwoFARequired) {
        setTwoFAPending(true);
        showToast({
          type: "info",
          message: "Check your email — we sent you a 6-digit sign-in code.",
        });
        return;
      }

      localStorage.setItem("auth", data.AuthToken.replace("Bearer ", ""));

      setupAuth();

      showToast({
        type: "success",
        message: "Login successful! Welcome back.",
      });
    },
  });

  const {
    post: postTwoFACode,
    loading: verifyingCode,
    error: twoFAError,
  } = usePost({
    url: API_POST_VERIFY_TWO_FA,
    callback: (data) => {
      if (!data || !data.AuthToken) return;

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

  useEffect(() => {
    if (twoFAError) {
      showToast({
        type: "danger",
        message: twoFAError,
      });
    }
  }, [twoFAError]);

  const handleCodeSubmit = (e) => {
    e.preventDefault();

    if (twoFACode.trim().length !== 6) {
      showToast({
        type: "danger",
        message: "Enter the 6-digit code from your email",
      });
      return;
    }

    postTwoFACode({ username: formData.username.trim(), code: twoFACode.trim() });
  };

  // once the code is on its way, the password form swaps for the code entry
  if (twoFAPending) {
    return (
      <form onSubmit={handleCodeSubmit} className='login-form-38jr'>
        <p className='text-center mb-4'>
          Enter the 6-digit code we sent to your email to finish signing in.
        </p>

        <div className='login-form__field'>
          <label htmlFor='two_fa_code' className='login-form__label'>
            Sign-in code
          </label>
          <Input
            onChange={(e) => setTwoFACode(e.target.value)}
            placeholder='6-digit code'
            className='mb-4'
            value={twoFACode}
            id='two_fa_code'
            maxLength={6}
            type='text'
            autoFocus
            required
          />
        </div>

        <Button
          type='submit'
          primary
          isLoading={verifyingCode}
          className='w-100 mb-1'
        >
          Verify &amp; Sign In
        </Button>
        <Button
          onClick={() => {
            setTwoFAPending(false);
            setTwoFACode("");
          }}
          className='w-100'
          type='button'
          secondary
        >
          Back
        </Button>
      </form>
    );
  }

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

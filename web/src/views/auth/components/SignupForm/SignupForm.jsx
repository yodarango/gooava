import { useAppContext } from "../../../context/appContextProvider";
import { Input, Button, Modal, Thumbnail, IfElse, AvatarPicker } from "@ds";
import React, { useState, useEffect } from "react";
import { API_POST_SIGNUP } from "@constants";
import { usePost } from "@utils";
import { avatars } from "@images";

// styles
import "./SignupForm.css";

export const SignupForm = () => {
  const { showToast, setupAuth } = useAppContext();
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const [formData, setFormData] = useState({
    confirmPassword: "",
    first_name: "",
    last_name: "",
    username: "",
    password: "",
    email: "",
    avatar: "",
  });

  const [showAvatarModal, setShowAvatarModal] = useState(false);

  const { post, loading, error } = usePost({
    url: API_POST_SIGNUP,
    callback: (data) => {
      if (!data) return;

      localStorage.setItem("auth", data.AuthToken.replace("Bearer ", ""));

      setupAuth();

      showToast({
        message: "Account created successfully! Welcome to Goilerplate.",
        type: "success",
      });
    },
  });

  const handleInputChange = (field) => (e) => {
    setFormData((prev) => ({
      ...prev,
      [field]: e.target.value,
    }));
  };

  const handleAvatarSave = (avatarPath) => {
    setFormData((prev) => ({
      ...prev,
      avatar: avatarPath,
    }));
    setShowAvatarModal(false);
  };

  const chosenAvater =
    avatars.find((avatar) => avatar.pathName === formData.avatar) || {};

  const validateForm = () => {
    if (!formData.first_name.trim()) {
      showToast({ type: "danger", message: "First name is required" });
      return false;
    }
    if (!formData.last_name.trim()) {
      showToast({ type: "danger", message: "Last name is required" });
      return false;
    }
    if (!formData.username.trim()) {
      showToast({ type: "danger", message: "Username is required" });
      return false;
    }
    if (!formData.email.trim()) {
      showToast({ type: "danger", message: "Email is required" });
      return false;
    }
    if (!formData.email.includes("@")) {
      showToast({
        type: "danger",
        message: "Please enter a valid email address",
      });
      return false;
    }
    if (!formData.password.trim()) {
      showToast({ type: "danger", message: "Password is required" });
      return false;
    }
    if (formData.password.length < 6) {
      showToast({
        type: "danger",
        message: "Password must be at least 6 characters long",
      });
      return false;
    }
    if (formData.password !== formData.confirmPassword) {
      showToast({ type: "danger", message: "Passwords do not match" });
      return false;
    }
    if (!formData.avatar.trim()) {
      showToast({ type: "danger", message: "Please select an avatar" });
      return false;
    }
    return true;
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    const { confirmPassword, ...signupData } = formData;
    post(signupData);
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
    <form onSubmit={handleSubmit} className='signup-form-hd4l'>
      <div>
        <div className='signup-form__field'>
          <label htmlFor='first_name'>First Name</label>
          <Input
            onChange={handleInputChange("first_name")}
            value={formData.first_name}
            placeholder='First name'
            className='mb-4'
            id='first_name'
            type='text'
            required
          />
        </div>

        <div className='signup-form__field'>
          <label htmlFor='last_name'>Last Name</label>
          <Input
            onChange={handleInputChange("last_name")}
            value={formData.last_name}
            placeholder='Last name'
            className='mb-4'
            id='last_name'
            type='text'
            required
          />
        </div>
      </div>

      <div className='signup-form__field'>
        <label htmlFor='username' className='signup-form__label'>
          Username
        </label>
        <Input
          onChange={handleInputChange("username")}
          placeholder='Choose a username'
          value={formData.username}
          className='mb-4'
          id='username'
          type='text'
          required
        />
      </div>

      <div className='signup-form__field'>
        <label htmlFor='email' className='signup-form__label'>
          Email
        </label>
        <Input
          onChange={handleInputChange("email")}
          placeholder='Enter your email'
          value={formData.email}
          className='mb-4'
          type='email'
          id='email'
          required
        />
      </div>

      <div className='signup-form__field'>
        <label htmlFor='password' className='signup-form__label'>
          Password
        </label>
        <div className='position-relative'>
          <Input
            onChange={handleInputChange("password")}
            placeholder='Create a password'
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

      <div className='signup-form__field'>
        <label htmlFor='confirmPassword' className='signup-form__label'>
          Confirm Password
        </label>
        <div className='position-relative'>
          <Input
            onChange={handleInputChange("confirmPassword")}
            placeholder='Confirm your password'
            value={formData.confirmPassword}
            id='confirmPassword'
            className='mb-4'
            type={showConfirmPassword ? "text" : "password"}
            required
          />
          <button
            type='button'
            className='password-toggle-btn position-absolute'
            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
          >
            <ion-icon
              name={showConfirmPassword ? "eye-off-outline" : "eye-outline"}
            ></ion-icon>
          </button>
        </div>
      </div>

      <div className='d-flex flex-column justify-content-center align-items-center'>
        <IfElse condition={formData.avatar && chosenAvater.image}>
          <div className='bg-zeta-20 w-100 p-4 rounded-2 mb-4 d-flex align-items-center justify-content-start gap-4'>
            <div
              className='signup-form-hd4l_avatar d-flex align-items-center justify-content-center flex-shrink-0'
              onClick={() => setShowAvatarModal(true)}
            >
              <Thumbnail
                alt={`Selected avatar: ${chosenAvater.name}`}
                src={chosenAvater.image}
                width='90%'
              />
            </div>
            <p className='signup-form-hd4l_avatar__name mb-0 opacity-60 text-center'>
              {chosenAvater.animal}
            </p>
          </div>
          <Button
            onClick={() => setShowAvatarModal(true)}
            className='w-100 mb-4'
            type='button'
            primary
          >
            Select avatar
          </Button>
        </IfElse>
      </div>

      <Button type='submit' success isLoading={loading} className='w-100'>
        Create Account
      </Button>

      <Modal
        onClose={() => setShowAvatarModal(false)}
        title='Choose Your Avatar'
        open={showAvatarModal}
        showWaves={false}
      >
        <AvatarPicker onSave={handleAvatarSave} isLoading={loading} />
      </Modal>
    </form>
  );
};

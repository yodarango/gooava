import { API_POST_CHANGE_PASSWORD, API_POST_UPDATE_PROFILE } from "@constants";
import { Input, Button, Modal, Thumbnail, AvatarPicker, IfElse } from "@ds";
import { useAppContext } from "../../context/appContextProvider";
import { useState, useEffect } from "react";
import { usePost } from "@utils";
import { avatars } from "@images";

// styles
import "./Layout.css";

export const Layout = () => {
  const { state, showToast, logout, setupAuth } = useAppContext();
  const { user } = state;

  const [showAvatarModal, setShowAvatarModal] = useState(false);
  const [showPasswords, setShowPasswords] = useState(false);

  const [formData, setFormData] = useState({
    first_name: user.first_name || "",
    last_name: user.last_name || "",
    username: user.username || "",
    email: user.email || "",
    avatar: user.avatar || "",
  });

  const [passwordData, setPasswordData] = useState({
    current_password: "",
    new_password: "",
    confirm_password: "",
  });

  const {
    post: saveProfile,
    loading: savingProfile,
    error: profileError,
  } = usePost({
    url: API_POST_UPDATE_PROFILE,
    callback: (data) => {
      if (!data) return;

      // the API returns a fresh token carrying the updated profile claims,
      // so the session (and the header avatar) reflects the changes at once
      if (data.AuthToken) {
        localStorage.setItem("auth", data.AuthToken.replace("Bearer ", ""));
        setupAuth();
      }

      showToast({ message: "Profile updated", type: "success" });
    },
  });

  const {
    post: changePassword,
    loading: changingPassword,
    error: passwordError,
  } = usePost({
    url: API_POST_CHANGE_PASSWORD,
    callback: (data) => {
      if (!data) return;
      setPasswordData({
        current_password: "",
        new_password: "",
        confirm_password: "",
      });
      showToast({ message: "Password changed", type: "success" });
    },
  });

  // surface API errors as toasts
  useEffect(() => {
    if (profileError) showToast({ type: "danger", message: profileError });
  }, [profileError]);

  useEffect(() => {
    if (passwordError) showToast({ type: "danger", message: passwordError });
  }, [passwordError]);

  const handleInputChange = (field) => (e) =>
    setFormData((prev) => ({ ...prev, [field]: e.target.value }));

  const handlePasswordInput = (field) => (e) =>
    setPasswordData((prev) => ({ ...prev, [field]: e.target.value }));

  const handleAvatarSave = (avatarPath) => {
    setFormData((prev) => ({ ...prev, avatar: avatarPath }));
    setShowAvatarModal(false);
  };

  const validateProfile = () => {
    if (!formData.first_name.trim() || !formData.last_name.trim()) {
      showToast({ type: "danger", message: "First and last name are required" });
      return false;
    }
    if (!formData.username.trim()) {
      showToast({ type: "danger", message: "Username is required" });
      return false;
    }
    if (!formData.email.trim() || !formData.email.includes("@")) {
      showToast({ type: "danger", message: "Please enter a valid email address" });
      return false;
    }
    if (!formData.avatar.trim()) {
      showToast({ type: "danger", message: "Please select an avatar" });
      return false;
    }
    return true;
  };

  const handleProfileSubmit = (e) => {
    e.preventDefault();
    if (!validateProfile()) return;
    saveProfile(formData);
  };

  const handlePasswordSubmit = (e) => {
    e.preventDefault();

    if (!passwordData.current_password) {
      showToast({ type: "danger", message: "Enter your current password" });
      return;
    }
    if (passwordData.new_password.length < 6) {
      showToast({
        type: "danger",
        message: "New password must be at least 6 characters long",
      });
      return;
    }
    if (passwordData.new_password !== passwordData.confirm_password) {
      showToast({ type: "danger", message: "New passwords do not match" });
      return;
    }

    changePassword({
      current_password: passwordData.current_password,
      new_password: passwordData.new_password,
    });
  };

  const chosenAvatar =
    avatars.find((avatar) => avatar.pathName === formData.avatar) || null;

  return (
    <div className='me-layout-9f3k'>
      <h1 className='me-layout-9f3k__title'>Profile</h1>

      {/* profile card */}
      <form onSubmit={handleProfileSubmit} className='bg-gamma rounded-3 p-4 mb-4'>
        <div className='d-flex flex-column align-items-center mb-4'>
          <button
            className='me-layout-9f3k__avatar-btn'
            onClick={() => setShowAvatarModal(true)}
            aria-label='Change avatar'
            type='button'
          >
            <IfElse condition={!!chosenAvatar}>
              <Thumbnail
                alt={`Selected avatar: ${chosenAvatar?.animal || ""}`}
                src={chosenAvatar?.image}
                width='90%'
              />
              <span className='me-layout-9f3k__avatar-fallback'>
                {(formData.first_name[0] || "?").toUpperCase()}
              </span>
            </IfElse>
          </button>
          <p className='opacity-60 mt-2 mb-0'>Tap to change your avatar</p>
        </div>

        <label className='me-layout-9f3k__label' htmlFor='first_name'>
          First Name
        </label>
        <Input
          onChange={handleInputChange("first_name")}
          value={formData.first_name}
          placeholder='First name'
          className='mb-4'
          id='first_name'
          type='text'
          required
        />

        <label className='me-layout-9f3k__label' htmlFor='last_name'>
          Last Name
        </label>
        <Input
          onChange={handleInputChange("last_name")}
          value={formData.last_name}
          placeholder='Last name'
          className='mb-4'
          id='last_name'
          type='text'
          required
        />

        <label className='me-layout-9f3k__label' htmlFor='username'>
          Username
        </label>
        <Input
          onChange={handleInputChange("username")}
          value={formData.username}
          placeholder='Username'
          className='mb-4'
          id='username'
          type='text'
          required
        />

        <label className='me-layout-9f3k__label' htmlFor='email'>
          Email
        </label>
        <Input
          onChange={handleInputChange("email")}
          value={formData.email}
          placeholder='Email'
          className='mb-4'
          id='email'
          type='email'
          required
        />

        <Button type='submit' success isLoading={savingProfile} className='w-100'>
          Save changes
        </Button>
      </form>

      {/* password card */}
      <form onSubmit={handlePasswordSubmit} className='bg-gamma rounded-3 p-4 mb-4'>
        <div className='d-flex align-items-center justify-content-between mb-4'>
          <h4 className='me-layout-9f3k__section-title m-0'>Change password</h4>
          <button
            className='me-layout-9f3k__password-toggle'
            onClick={() => setShowPasswords(!showPasswords)}
            type='button'
          >
            {showPasswords ? "Hide" : "Show"}
          </button>
        </div>

        <label className='me-layout-9f3k__label' htmlFor='current_password'>
          Current password
        </label>
        <Input
          onChange={handlePasswordInput("current_password")}
          value={passwordData.current_password}
          placeholder='Current password'
          type={showPasswords ? "text" : "password"}
          className='mb-4'
          id='current_password'
          required
        />

        <label className='me-layout-9f3k__label' htmlFor='new_password'>
          New password
        </label>
        <Input
          onChange={handlePasswordInput("new_password")}
          value={passwordData.new_password}
          placeholder='New password'
          type={showPasswords ? "text" : "password"}
          className='mb-4'
          id='new_password'
          required
        />

        <label className='me-layout-9f3k__label' htmlFor='confirm_password'>
          Confirm new password
        </label>
        <Input
          onChange={handlePasswordInput("confirm_password")}
          value={passwordData.confirm_password}
          placeholder='Confirm new password'
          type={showPasswords ? "text" : "password"}
          className='mb-4'
          id='confirm_password'
          required
        />

        <Button type='submit' primary isLoading={changingPassword} className='w-100'>
          Change password
        </Button>
      </form>

      {/* logout — ProtectedRoute redirects to /auth once logged out */}
      <Button danger className='w-100' onClick={logout} type='button'>
        Log out
      </Button>

      <Modal
        onClose={() => setShowAvatarModal(false)}
        title='Choose Your Avatar'
        open={showAvatarModal}
        showWaves={false}
      >
        <AvatarPicker onSave={handleAvatarSave} isLoading={savingProfile} />
      </Modal>
    </div>
  );
};

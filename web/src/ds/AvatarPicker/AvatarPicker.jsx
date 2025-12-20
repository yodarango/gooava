import { Button, Modal, Thumbnail } from "@ds";
import { useEffect, useState } from "react";
import { avatars } from "@images";

// styles
import "./AvatarPicker.css";

export const AvatarPicker = (props) => {
  const { onSave, isLoading, onSuccess, withConfirmation } = props;

  const [showConfirmationModal, setShowConfirmationModal] = useState(false);
  const [currentAvatar, setCurrentAvatar] = useState("");

  const handleSave = () => {
    onSave(currentAvatar);
  };

  const handleAvatarClick = (animal) => {
    setShowConfirmationModal(true);
    setCurrentAvatar(animal);
  };

  useEffect(() => {
    if (onSuccess) {
      setShowConfirmationModal(false);
    }
  }, [onSuccess]);

  return (
    <>
      <Modal
        title='You are about to change your avatar, Please confirm'
        onClose={() => setShowConfirmationModal(false)}
        open={showConfirmationModal}
        showWaves={false}
        zIndex={17}
      >
        <div className='shrood-comment-card-47kd__actions d-flex align-items-center justify-content-center w-100 gap-4 mt-4'>
          <Button
            onClick={() => setShowConfirmationModal(false)}
            className='w-50'
            secondary
          >
            Cancel
          </Button>
          <Button
            loading={isLoading}
            onClick={handleSave}
            className='w-50'
            primary
          >
            Save
          </Button>
        </div>
      </Modal>
      <div className='avatar-picker-01lp d-flex align-items-start justify-content-center flex-wrap flex-wrap gap-4'>
        {avatars.map((avatar) => (
          <div
            className='avatar-picker-01lp__avatar d-inline-flex align-items-center justify-content-center flex-shrink-0 p-4'
            key={avatar.id}
          >
            <div
              className='d-inline-flex align-items-center justify-content-center flex-shrink-0 bg-gamma'
              onClick={() => {
                if (withConfirmation) handleAvatarClick(avatar.pathName);
                else onSave(avatar.pathName);
              }}
            >
              <Thumbnail
                alt="avatar for user's profile"
                className='rounded-4'
                src={avatar.image}
                width={"100%"}
              />
            </div>
          </div>
        ))}
      </div>
    </>
  );
};

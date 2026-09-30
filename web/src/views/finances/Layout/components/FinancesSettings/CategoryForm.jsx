import { useAppContext } from "../../../../context/appContextProvider";
import { API_POST_EXPENSE_CATEGORIES } from "@constants";
import { Button, Input, TextArea } from "@ds";
import { useEffect, useState } from "react";
import { usePost } from "@utils";

// styles
import "./CategoryForm.css";

const DEFAULT_COLOR = "#f24c66"; // --dr-delta

/*********************************************************************************************************
 * Form to add an expense category: label, description and a color picked with the native color input.
 * Expenses link to categories by the label string.
 * ******************************************************************************************************
 */
export const CategoryForm = (props) => {
  const { onDone, onCancel } = props;
  const { showToast } = useAppContext();

  const [label, setLabel] = useState("");
  const [description, setDescription] = useState("");
  const [color, setColor] = useState(DEFAULT_COLOR);

  const create = usePost({
    url: API_POST_EXPENSE_CATEGORIES,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Category added", type: "success" });
      if (onDone) onDone();
    },
  });

  // surface request errors as toasts
  useEffect(() => {
    if (create.error) showToast({ message: String(create.error), type: "danger" });
  }, [create.error]);

  const handleSubmit = (e) => {
    e.preventDefault();

    if (label.trim() === "") {
      showToast({ message: "Please enter a label", type: "danger" });
      return;
    }

    create.post({
      label: label.trim(),
      description: description.trim(),
      color,
    });
  };

  return (
    <form className='category-form-9qw4' onSubmit={handleSubmit}>
      <div className='category-form-9qw4__field'>
        <label className='category-form-9qw4__label' htmlFor='category-label'>
          Label
        </label>
        <Input
          onChange={(e) => setLabel(e.target.value)}
          placeholder='e.g. Groceries'
          id='category-label'
          value={label}
          type='text'
          required
        />
      </div>

      <div className='category-form-9qw4__field'>
        <label className='category-form-9qw4__label' htmlFor='category-description'>
          Description
        </label>
        <TextArea
          onChange={(e) => setDescription(e.target.value)}
          placeholder='What belongs in this category?'
          id='category-description'
          value={description}
          minRows={2}
          maxRows={5}
        />
      </div>

      <div className='category-form-9qw4__field'>
        <label className='category-form-9qw4__label' htmlFor='category-color'>
          Color
        </label>
        <div className='category-form-9qw4__color'>
          <input
            onChange={(e) => setColor(e.target.value)}
            className='category-form-9qw4__color-input'
            id='category-color'
            value={color}
            type='color'
          />
          <span className='category-form-9qw4__color-hex'>{color}</span>
        </div>
      </div>

      <div className='category-form-9qw4__actions'>
        <Button primary type='submit' isLoading={create.loading} className='w-100'>
          Add category
        </Button>
        {onCancel && (
          <Button secondary type='button' onClick={onCancel}>
            Cancel
          </Button>
        )}
      </div>
    </form>
  );
};

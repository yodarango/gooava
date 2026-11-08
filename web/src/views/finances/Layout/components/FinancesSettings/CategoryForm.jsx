import { useAppContext } from "../../../../context/appContextProvider";
import {
  API_POST_EXPENSE_CATEGORIES_UPDATE,
  API_POST_EXPENSE_CATEGORIES,
} from "@constants";
import { Button, Input, TextArea } from "@ds";
import { useEffect, useState } from "react";
import { usePost } from "@utils";

// styles
import "./CategoryForm.css";

const DEFAULT_COLOR = "#f24c66"; // --dr-delta

/*********************************************************************************************************
 * Form to add or edit an expense category: label, description and a color picked with
 * the native color input. With the `category` prop it edits (and a rename re-points the
 * expenses linked by the old label); without it, it creates. Expenses link to categories
 * by the label string.
 * ******************************************************************************************************
 */
export const CategoryForm = (props) => {
  const { onDone, onCancel, category } = props;
  const editing = !!category;
  const { showToast } = useAppContext();

  const [label, setLabel] = useState(category?.label || "");
  const [description, setDescription] = useState(category?.description || "");
  const [color, setColor] = useState(category?.color || DEFAULT_COLOR);

  const create = usePost({
    url: API_POST_EXPENSE_CATEGORIES,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Category added", type: "success" });
      if (onDone) onDone();
    },
  });

  const update = usePost({
    url: API_POST_EXPENSE_CATEGORIES_UPDATE,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Category updated", type: "success" });
      if (onDone) onDone();
    },
  });

  // surface request errors as toasts
  const request = editing ? update : create;
  useEffect(() => {
    if (request.error) showToast({ message: String(request.error), type: "danger" });
  }, [request.error]);

  const handleSubmit = (e) => {
    e.preventDefault();

    if (label.trim() === "") {
      showToast({ message: "Please enter a label", type: "danger" });
      return;
    }

    const payload = {
      label: label.trim(),
      description: description.trim(),
      color,
    };

    if (editing) {
      update.post({ id: category.id, ...payload });
    } else {
      create.post(payload);
    }
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
        <Button primary type='submit' isLoading={request.loading} className='w-100'>
          {editing ? "Save changes" : "Add category"}
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

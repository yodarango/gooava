import { useAppContext } from "../../../../context/appContextProvider";
import {
  API_POST_EXPENSE_CATEGORIES_DELETE,
  API_GET_EXPENSE_CATEGORIES,
} from "@constants";
import { CategoryForm } from "./CategoryForm";
import { Pencil, Plus, Trash2, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useEffect, useState } from "react";
import { useGet, usePost } from "@utils";

// styles
import "./CategoriesPane.css";

/*********************************************************************************************************
 * The Categories tab: a title row ("Categories" + plus icon that opens the add form) above the list
 * of the user's categories. Expenses link to categories by the label string.
 * ******************************************************************************************************
 */
export const CategoriesPane = () => {
  const { showToast } = useAppContext();

  const [formOpen, setFormOpen] = useState(false);
  const [editing, setEditing] = useState(null); // category being edited, null = add mode
  const [refreshKey, setRefreshKey] = useState(0);

  const categories = useGet({
    url: API_GET_EXPENSE_CATEGORIES,
    dependencies: [refreshKey],
  });

  const deleteCategory = usePost({
    url: API_POST_EXPENSE_CATEGORIES_DELETE,
    callback: (data) => {
      if (!data) return;
      showToast({ message: "Category deleted", type: "success" });
      setRefreshKey((key) => key + 1);
    },
  });

  // surface delete errors as toasts
  useEffect(() => {
    if (deleteCategory.error) {
      showToast({ message: String(deleteCategory.error), type: "danger" });
    }
  }, [deleteCategory.error]);

  // plus opens the add form (or closes it if already open in add mode)
  const handlePlus = () => {
    if (formOpen && !editing) {
      setFormOpen(false);
      return;
    }
    setEditing(null);
    setFormOpen(true);
  };

  const handleEdit = (category) => {
    setEditing(category);
    setFormOpen(true);
  };

  const handleDelete = (category) => {
    if (!window.confirm(`Delete the "${category.label}" category?`)) return;
    deleteCategory.post({ id: category.id });
  };

  const handleFormDone = () => {
    setFormOpen(false);
    setEditing(null);
    setRefreshKey((key) => key + 1);
  };

  const categoryList = Array.isArray(categories.data) ? categories.data : [];

  return (
    <div className='categories-pane-7zx2'>
      <div className='categories-pane-7zx2__header'>
        <h3 className='categories-pane-7zx2__title'>Categories</h3>
        <button
          className='categories-pane-7zx2__add'
          onClick={handlePlus}
          aria-label={formOpen ? "Close form" : "Add category"}
          type='button'
        >
          <MorphIcon
            icon={formOpen ? X : Plus}
            label='Add category'
            size={20}
          />
        </button>
      </div>

      {formOpen && (
        <div className='categories-pane-7zx2__form'>
          <CategoryForm
            key={editing ? editing.id : "new"}
            category={editing}
            onDone={handleFormDone}
            onCancel={handleFormDone}
          />
        </div>
      )}

      <div className='categories-pane-7zx2__list'>
        {categories.loading && categoryList.length === 0 ? (
          <p className='categories-pane-7zx2__empty'>Loading…</p>
        ) : categoryList.length === 0 ? (
          <p className='categories-pane-7zx2__empty'>
            No categories yet — tap + to add one.
          </p>
        ) : (
          <ul>
            {categoryList.map((category) => (
              <li key={category.id} className='categories-pane-7zx2__row'>
                <span
                  className='categories-pane-7zx2__swatch'
                  style={{ backgroundColor: category.color }}
                />
                <div className='categories-pane-7zx2__details'>
                  <span className='categories-pane-7zx2__label'>{category.label}</span>
                  {category.description && (
                    <span className='categories-pane-7zx2__meta'>
                      {category.description}
                    </span>
                  )}
                </div>
                <span className='categories-pane-7zx2__hex'>{category.color}</span>

                <div className='categories-pane-7zx2__actions'>
                  <button
                    className='categories-pane-7zx2__action edit'
                    onClick={() => handleEdit(category)}
                    aria-label={`Edit ${category.label}`}
                    title='Edit category'
                    type='button'
                  >
                    <MorphIcon icon={Pencil} label='Edit' size={15} />
                  </button>
                  <button
                    className='categories-pane-7zx2__action delete'
                    onClick={() => handleDelete(category)}
                    aria-label={`Delete ${category.label}`}
                    disabled={deleteCategory.loading}
                    title='Delete category'
                    type='button'
                  >
                    <MorphIcon icon={Trash2} label='Delete' size={15} />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

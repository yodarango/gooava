import { API_GET_EXPENSE_CATEGORIES } from "@constants";
import { CategoryForm } from "./CategoryForm";
import { Plus, X } from "lucide"; // data, not components
import { MorphIcon } from "morphicons/react";
import { useState } from "react";
import { useGet } from "@utils";

// styles
import "./CategoriesPane.css";

/*********************************************************************************************************
 * The Categories tab: a title row ("Categories" + plus icon that opens the add form) above the list
 * of the user's categories. Expenses link to categories by the label string.
 * ******************************************************************************************************
 */
export const CategoriesPane = () => {
  const [formOpen, setFormOpen] = useState(false);
  const [refreshKey, setRefreshKey] = useState(0);

  const categories = useGet({
    url: API_GET_EXPENSE_CATEGORIES,
    dependencies: [refreshKey],
  });

  const handleFormDone = () => {
    setFormOpen(false);
    setRefreshKey((key) => key + 1);
  };

  const categoryList = Array.isArray(categories.data) ? categories.data : [];

  return (
    <div className='categories-pane-7zx2'>
      <div className='categories-pane-7zx2__header'>
        <h3 className='categories-pane-7zx2__title'>Categories</h3>
        <button
          className='categories-pane-7zx2__add'
          onClick={() => setFormOpen((open) => !open)}
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
          <CategoryForm onDone={handleFormDone} onCancel={handleFormDone} />
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
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
};

/*********************************************************************************************************
 * This is the entry point of the application. It displays the different kinds of functionalities and
 * features that the application provides. Perhaps at some point might be a good idea to fetch all
 * these options from the backend, sinc they provide dynamic data like the number of terms,
 * and different card sets or features. Otherwise I will need to keep changing the view
 * code every time I add a new feature or card set.
 * ******************************************************************************************************
 */

import { IntroSection } from "./components/IntroSection/IntroSection";

// styles
import "./Layout.css";

export const Layout = () => {
  return (
    <div className='index-layout-container'>
      <IntroSection />
    </div>
  );
};

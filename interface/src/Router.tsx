import { BrowserRouter, Route, Routes } from "react-router-dom"
import SearchPage from "./pages/SearchPage.tsx"
import PageNotFound from "./components/PageNotFound.tsx"

const Router = () => {
  return (
    <BrowserRouter future={{ v7_startTransition: true }}>
      <Routes>
        <Route path={"/"} element={<SearchPage />} />
        <Route path={"*"} element={<PageNotFound />} />
      </Routes>
    </BrowserRouter>
  )
}

export default Router

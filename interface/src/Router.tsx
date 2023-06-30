import { BrowserRouter, Route, Routes } from 'react-router-dom'
import SearchPage from "./pages/SearchPage.tsx"

const Router = () =>  {
    return (
        <BrowserRouter future={{ v7_startTransition: true}}>
            <Routes>
                <Route path={"/"} element={<div>Hello World</div>}/>
                <Route path={"/search"} element={<SearchPage/>}/>
            </Routes>
        </BrowserRouter>
    )
}

export default Router
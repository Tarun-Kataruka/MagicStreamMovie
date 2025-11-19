import React from "react";
import Home from "./components/home/Home";
import Header from "./components/header/Header";
import Register from "./components/register/Register";
import Login from "../src/components/login/Login";
import { Route, Routes } from "react-router-dom";
import Layout from "./components/Layout";
import RequireAuth from "./components/RequiredAuth";
import Recommended from "./components/recommended/Recommended";

import "./App.css";

function App() {
  return (
    <>
      <Header />
      <Routes path="/" element={<Layout />}>
        <Route path="/" element={<Home />} />
        <Route path="/register" element={<Register />} />
        <Route path="/login" element={<Login />} />
        <Route element={<RequireAuth />}>
          <Route path="/recommended" element={<Recommended />} />
        </Route>
      </Routes>
    </>
  );
}

export default App;

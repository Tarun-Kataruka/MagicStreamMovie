import React from "react";
import Home from "./components/home/Home";
import Header from "./components/header/Header";
import Register from "./components/register/Register";
import Login from "../src/components/login/Login";
import { Route, Routes, useNavigate } from "react-router-dom";
import Layout from "./components/Layout";
import RequireAuth from "./components/RequiredAuth";
import Recommended from "./components/recommended/Recommended";
import Review from "./components/review/Review";
import "./App.css";

function App() {
  const navigate = useNavigate();
  const updateMovieReview = (imdb_id) => {
    navigate(`/review/${imdb_id}`);
  };
  return (
    <>
      <Header />
      <Routes path="/" element={<Layout />}>
        <Route
          path="/"
          element={<Home updateMovieReview={updateMovieReview} />}
        />
        <Route path="/register" element={<Register />} />
        <Route path="/login" element={<Login />} />
        <Route element={<RequireAuth />}>
          <Route path="/recommended" element={<Recommended />} />
          <Route path="/review/:imdb_id" element={<Review />} />
        </Route>
      </Routes>
    </>
  );
}

export default App;

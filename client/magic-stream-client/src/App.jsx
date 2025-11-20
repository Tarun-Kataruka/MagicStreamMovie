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
import useAuth from "./hook/useAuth";
import axiosClient from "./api/axiosConfig";
import StreamMovie from "./components/stream/StreamMovie";
import "./App.css";

function App() {
  const navigate = useNavigate();
  const { auth, setAuth } = useAuth();

  const updateMovieReview = (imdb_id) => {
    navigate(`/review/${imdb_id}`);
  };

  const handleLogout = async () => {
    try {
        const response = await axiosClient.post("/logout",{user_id: auth.user_id});
        console.log(response.data);
        setAuth(null);
        console.log('User logged out'); 
      } catch (error) {
        console.error('Error logging out:', error);
      } 
  };

  return (
    <>
      <Header handleLogout={handleLogout} />
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
          <Route path="/stream/:yt_id" element={<StreamMovie />} />
        </Route>
      </Routes>
    </>
  );
}

export default App;

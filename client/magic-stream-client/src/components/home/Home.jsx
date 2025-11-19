import React from "react";
import { useState, useEffect } from "react";
import axiosConfig from "../../api/axiosConfig";
import Movies from "../movies/Movies";

const Home = ({updateMovieReview}) => {
  const [movies, setMovies] = useState([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState();

  useEffect(() => {
    const fetchMovies = async () => {
      setLoading(true);
      setMessage("");
      try {
        const response = await axiosConfig.get("/movies");
        setMovies(response.data);
        if (response.data.length === 0) {
          setMessage("No movies available");
        }
      } catch (error) {
        console.error("Error fetching movies:", error);
        setMessage("Failed to fetch movies");
      } finally {
        setLoading(false);
      }
    };
    fetchMovies();
  }, []);
  return (
    <>
      {loading ? (
        <h2>Loading...</h2>
      ) : (
        <Movies movies={movies} updateMovieReview={updateMovieReview} message={message} />
      )}
    </>
  );
};

export default Home;

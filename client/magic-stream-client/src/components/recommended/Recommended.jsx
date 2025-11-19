import useAxiosPrivate from "../../hook/useAxiosPrivate";
import Movies from "../movies/Movies";
import { useEffect, useState } from "react";

const Recommended = () => {
  const [movies, setMovies] = useState([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const axiosPrivate = useAxiosPrivate();

  useEffect(() => {
    const fetchRecommended = async () => {
      setLoading(true);
      setMessage("");
      try {
        const response = await axiosPrivate.get("/recommendedmovies");
        setMovies(response.data);
      } catch (error) {
        console.error("Error fetching recommended movies:", error);
        setMessage("Failed to load recommended movies.");
      } finally {
        setLoading(false);
      }
    };
    fetchRecommended();
  }, [axiosPrivate]);

  return (
    <>
      {loading ? (
        <h2>Loading....</h2>
      ) : (
        <Movies movies={movies} message={message} />
      )}
    </>
  );
};

export default Recommended;

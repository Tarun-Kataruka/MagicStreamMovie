import React from "react";
import Button from "react-bootstrap/Button";
const Movie = ({ movie, updateMoviewReview }) => {
  return (
    <div className="col-md-4 mb-4">
      <div className="card h-100 shadow-sm">
        <div style={{ position: "relative" }}>
          <img
            src={movie.poster_path}
            alt={movie.title}
            className="card-img-top"
            style={{
              objectFit: "contain",
              height: "250px",
              width: "100%",
            }}
          ></img>
        </div>
        <div className="card-body d-flex flex-column">
          <h5 className="card-title">{movie.title}</h5>
          <p className="card-text mb-2">{movie.imdb_id}</p>
        </div>
        {movie.ranking?.ranking_name && (
          <span className="badge bg-dark m-3 p-2" style={{ fontSize: "1rem" }}>
            {movie.ranking.ranking_name}
          </span>
        )}
        {updateMoviewReview && (
          <Button
            variant="outline-info"
            onClick={(e) => {
              e.preventDefault();
              updateMoviewReview(movie.imdb_id);
            }}
            className="m-3"
          >
            Review
          </Button>
        )}
      </div>
    </div>
  );
};
export default Movie;

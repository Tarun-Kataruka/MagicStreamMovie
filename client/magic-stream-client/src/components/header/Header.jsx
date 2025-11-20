import Button from "react-bootstrap/Button";
import Container from "react-bootstrap/Container";
import Nav from "react-bootstrap/Nav";
import Navbar from "react-bootstrap/Navbar";
import { useNavigate, NavLink, Link } from "react-router-dom";
import useAuth from "../../hook/useAuth";

const Header = ({handleLogout}) => {
  const navigate = useNavigate();
  const { auth } = useAuth();
  return (
    <Navbar
      bg="dark"
      variant="dark"
      expand="lg"
      stick="top"
      className="shadow-sm"
    >
      <Container>
        <Navbar.Brand as={Link} to="/">
          Magic Stream
        </Navbar.Brand>
        <Navbar.Toggle aria-controls="basic-navbar-nav" />
        <Navbar.Collapse>
          <Nav className="me-auto">
            <Nav.Link as={NavLink} to="/">
              Home
            </Nav.Link>
            <Nav.Link as={NavLink} to="/recommended">
              Recommended
            </Nav.Link>
          </Nav>
          <Nav className="ms-auto align-items-center">
            {auth ? (
              <>
                <span>
                  Hello, <strong>{auth.first_name}</strong>
                </span>
                <Button variant="outline-light" size="sm" onClick={handleLogout}>
                  Logout
                </Button>
              </>
            ) : (
              <>
                <Button
                  variant="outline-light"
                  size="sm"
                  className="me-2"
                  onClick={() => navigate("/login")}
                >
                  Login
                </Button>
                <Button
                  variant="outline-light"
                  size="sm"
                  onClick={() => navigate("/register")}
                >
                  Register
                </Button>
              </>
            )}
          </Nav>
        </Navbar.Collapse>
      </Container>
    </Navbar>
  );
};

export default Header;

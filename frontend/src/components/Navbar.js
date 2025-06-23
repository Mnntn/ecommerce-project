import React from 'react';
import { Link as RouterLink } from 'react-router-dom';
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Container,
  IconButton,
  Badge,
} from '@mui/material';
import ShoppingCartIcon from '@mui/icons-material/ShoppingCart';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import HomeIcon from '@mui/icons-material/Home';
import { useCartState } from '../context/CartContext';

function Navbar() {
  const { items } = useCartState();
  const totalItems = items.reduce((sum, item) => sum + item.quantity, 0);

  return (
    <AppBar position="static">
      <Container maxWidth="xl">
        <Toolbar>
          <Button
            color="inherit"
            component={RouterLink}
            to="/"
            startIcon={<HomeIcon />}
            sx={{ mr: 2 }}
          >
            <Typography variant="h6">E-Commerce</Typography>
          </Button>

          <Button
            color="inherit"
            component={RouterLink}
            to="/account"
            startIcon={<AccountBalanceIcon />}
          >
            Account
          </Button>

          <div style={{ flexGrow: 1 }} />

          <IconButton
            color="inherit"
            component={RouterLink}
            to="/cart"
            aria-label="show cart items"
          >
            <Badge badgeContent={totalItems} color="error">
              <ShoppingCartIcon />
            </Badge>
          </IconButton>
        </Toolbar>
      </Container>
    </AppBar>
  );
}

export default Navbar; 
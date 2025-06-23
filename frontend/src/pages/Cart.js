import React, { useState } from 'react';
import {
  Container,
  Typography,
  List,
  ListItem,
  ListItemText,
  Divider,
  Button,
  Box,
  TextField,
  Paper,
  Alert,
  CircularProgress
} from '@mui/material';
import { useCartState, useCartDispatch } from '../context/CartContext';
import { useNavigate } from 'react-router-dom';

function Cart() {
  const { items } = useCartState();
  const dispatch = useCartDispatch();
  const navigate = useNavigate();
  
  const [userId, setUserId] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [userBalance, setUserBalance] = useState(null);
  const [checkingBalance, setCheckingBalance] = useState(false);

  const totalPrice = items.reduce((sum, item) => sum + item.price * item.quantity, 0);

  const checkBalance = async () => {
    if (!userId) {
      setError('Please enter a User ID.');
      return;
    }
    
    setCheckingBalance(true);
    setError('');
    setUserBalance(null);
    
    try {
      const res = await fetch(`/api/payment/accounts/${userId}`);
      if (!res.ok) {
        throw new Error('User account not found. Please create an account first.');
      }
      const account = await res.json();
      setUserBalance(account.balance);
      
      if (account.balance < totalPrice) {
        setError(`Insufficient funds. Your balance: $${account.balance.toFixed(2)}, Required: $${totalPrice.toFixed(2)}`);
      }
    } catch (err) {
      setError(err.message || 'Failed to check balance.');
    } finally {
      setCheckingBalance(false);
    }
  };

  const handleCheckout = async () => {
    if (!userId) {
      setError('Please enter a User ID.');
      return;
    }
    
    setError('');
    setLoading(true);

    const orderRequest = {
      user_id: userId,
      items: items.map(item => ({
        product_id: item.id,
        quantity: item.quantity,
      })),
    };

    try {
      // Create the order - payment will be processed automatically via Kafka
      const orderRes = await fetch('/api/orders', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(orderRequest),
      });

      if (!orderRes.ok) {
        const errData = await orderRes.json();
        throw new Error(errData.message || 'Failed to create order.');
      }

      const order = await orderRes.json();
      
      // Order created successfully - payment processing happens automatically
      dispatch({ type: 'CLEAR_CART' });
      setUserBalance(null);
      
      // Show success message
      setError('');
      alert(`Order created successfully! Order ID: ${order.id}. Payment will be processed automatically.`);
      
      navigate('/account'); // Redirect to account page to view orders

    } catch (err) {
      setError(err.message || 'An unexpected error occurred during checkout.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom>
        Your Cart
      </Typography>
      <Paper elevation={3} sx={{ p: 2 }}>
        {items.length === 0 ? (
          <Typography>Your cart is empty.</Typography>
        ) : (
          <List>
            {items.map((item) => (
              <React.Fragment key={item.id}>
                <ListItem>
                  <ListItemText
                    primary={item.name}
                    secondary={`Quantity: ${item.quantity}`}
                  />
                  <Typography variant="body1">
                    ${(item.price * item.quantity).toFixed(2)}
                  </Typography>
                </ListItem>
                <Divider />
              </React.Fragment>
            ))}
            <ListItem>
              <ListItemText primary="Total" />
              <Typography variant="h6">
                ${totalPrice.toFixed(2)}
              </Typography>
            </ListItem>
          </List>
        )}
      </Paper>
      {items.length > 0 && (
        <Box mt={4}>
          <Typography variant="h6">Checkout</Typography>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          
          <TextField
            label="User UUID"
            variant="outlined"
            fullWidth
            value={userId}
            onChange={(e) => {
              setUserId(e.target.value);
              setUserBalance(null); // Reset balance when user ID changes
            }}
            sx={{ mb: 2 }}
            placeholder="Enter user UUID (e.g., 5e833693-2c14-40f9-95a1-a278148366aa)"
          />
          
          <Box display="flex" gap={2} mb={2}>
            <Button
              variant="outlined"
              onClick={checkBalance}
              disabled={checkingBalance || !userId}
              sx={{ minWidth: 120 }}
            >
              {checkingBalance ? <CircularProgress size={24} /> : 'Check Balance'}
            </Button>
            {userBalance !== null && (
              <Typography variant="body1" sx={{ alignSelf: 'center' }}>
                Balance: ${userBalance.toFixed(2)}
              </Typography>
            )}
          </Box>
          
          <Button
            variant="contained"
            color="primary"
            size="large"
            fullWidth
            onClick={handleCheckout}
            disabled={loading || !userId}
          >
            {loading ? <CircularProgress size={24} /> : 'Place Order'}
          </Button>
        </Box>
      )}
    </Container>
  );
}

export default Cart; 
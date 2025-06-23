import React, { useState, useEffect } from 'react';
import {
  Container,
  Typography,
  TextField,
  Button,
  Box,
  Card,
  CardContent,
  Alert,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  Divider,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper
} from '@mui/material';

function Account() {
  const [userId, setUserId] = useState('');
  const [account, setAccount] = useState(null);
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [users, setUsers] = useState([]);
  const [accounts, setAccounts] = useState({});
  const [newUserName, setNewUserName] = useState('');
  const [creating, setCreating] = useState(false);
  const [depositAmount, setDepositAmount] = useState('');
  const [depositing, setDepositing] = useState(false);
  const [success, setSuccess] = useState('');
  const [orders, setOrders] = useState([]);
  const [orderSearchUserId, setOrderSearchUserId] = useState('');
  const [orderSearchOrderId, setOrderSearchOrderId] = useState('');
  const [ordersLoading, setOrdersLoading] = useState(false);
  const [ordersError, setOrdersError] = useState('');

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      const res = await fetch('/api/payment/users');
      if (!res.ok) throw new Error('Failed to fetch users');
      const usersData = await res.json();
      setUsers(usersData);
      // Получаем счета для всех пользователей
      const accs = {};
      await Promise.all(usersData.map(async (u) => {
        const accRes = await fetch(`/api/payment/accounts/${u.id}`);
        if (accRes.ok) {
          accs[u.id] = await accRes.json();
        }
      }));
      setAccounts(accs);
    } catch (err) {
      setError(err.message || 'Error fetching users');
    }
  };

  const handleFetchAccount = async () => {
    setLoading(true);
    setError('');
    setAccount(null);
    setUser(null);
    try {
      const userRes = await fetch(`/api/payment/users/${userId}`);
      if (!userRes.ok) throw new Error('User not found');
      const userData = await userRes.json();
      setUser(userData);
      const accRes = await fetch(`/api/payment/accounts/${userId}`);
      if (!accRes.ok) throw new Error('Account not found');
      const accData = await accRes.json();
      setAccount(accData);
    } catch (err) {
      setError(err.message || 'Error fetching account');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateUser = async () => {
    setCreating(true);
    setError('');
    try {
      const res = await fetch('/api/payment/users', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newUserName })
      });
      if (!res.ok) throw new Error('Failed to create user');
      setNewUserName('');
      await fetchUsers();
    } catch (err) {
      setError(err.message || 'Error creating user');
    } finally {
      setCreating(false);
    }
  };

  const handleCreateOrDeposit = async () => {
    setDepositing(true);
    setError('');
    setSuccess('');
    try {
      // Проверяем, есть ли счет
      let accRes = await fetch(`/api/payment/accounts/${userId}`);
      if (!accRes.ok) {
        // Если счета нет — создаём
        const createRes = await fetch('/api/payment/accounts', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'X-User-ID': userId
          }
        });
        if (!createRes.ok) throw new Error('Не удалось создать счет');
        accRes = await fetch(`/api/payment/accounts/${userId}`);
        if (!accRes.ok) throw new Error('Не удалось получить счет после создания');
      }
      // Если введена сумма — пополняем
      if (depositAmount && parseFloat(depositAmount) > 0) {
        const depRes = await fetch(`/api/payment/accounts/${userId}/deposit`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ amount: parseFloat(depositAmount) })
        });
        if (!depRes.ok) throw new Error('Не удалось пополнить счет');
        setSuccess('Счет успешно создан и пополнен!');
      } else {
        setSuccess('Счет успешно создан!');
      }
      setDepositAmount('');
      await handleFetchAccount();
      await fetchUsers();
    } catch (err) {
      setError(err.message || 'Ошибка при создании/пополнении счета');
    } finally {
      setDepositing(false);
    }
  };

  const fetchAllOrders = async () => {
    setOrdersLoading(true);
    setOrdersError('');
    try {
      const res = await fetch('/api/orders');
      if (!res.ok) throw new Error('Не удалось получить список заказов');
      const data = await res.json();
      setOrders(Array.isArray(data) ? data : []);
    } catch (err) {
      setOrdersError(err.message || 'Ошибка при получении заказов');
    } finally {
      setOrdersLoading(false);
    }
  };

  const fetchOrdersByUserId = async () => {
    if (!orderSearchUserId) return;
    setOrdersLoading(true);
    setOrdersError('');
    try {
      const res = await fetch(`/api/orders/user/${orderSearchUserId}`);
      if (!res.ok) throw new Error('Не удалось получить заказы пользователя');
      const data = await res.json();
      setOrders(Array.isArray(data) ? data : []);
    } catch (err) {
      setOrdersError(err.message || 'Ошибка при получении заказов пользователя');
    } finally {
      setOrdersLoading(false);
    }
  };

  const fetchOrderById = async () => {
    if (!orderSearchOrderId) return;
    setOrdersLoading(true);
    setOrdersError('');
    try {
      const res = await fetch(`/api/orders/${orderSearchOrderId}`);
      if (!res.ok) throw new Error('Не удалось получить заказ по id');
      const data = await res.json();
      setOrders(data ? [data] : []);
    } catch (err) {
      setOrdersError(err.message || 'Ошибка при получении заказа по id');
    } finally {
      setOrdersLoading(false);
    }
  };

  return (
    <Container maxWidth="md" sx={{ mt: 4 }}>
      <Typography variant="h4" gutterBottom align="center">
        User Accounts
      </Typography>
      <Box display="flex" gap={2} mb={3}>
        <TextField
          label="Создать/Пополнить счёт по user_id"
          variant="outlined"
          value={userId}
          onChange={e => setUserId(e.target.value)}
          fullWidth
        />
        <TextField
          label="Сумма для пополнения"
          type="number"
          variant="outlined"
          value={depositAmount}
          onChange={e => setDepositAmount(e.target.value)}
          sx={{ minWidth: 180 }}
        />
        <Button
          variant="contained"
          color="primary"
          onClick={handleCreateOrDeposit}
          disabled={depositing || !userId}
        >
          {depositing ? <CircularProgress size={24} /> : 'Создать/Пополнить'}
        </Button>
      </Box>
      {error && <Alert severity="error">{error}</Alert>}
      {success && <Alert severity="success">{success}</Alert>}
      <Box mb={3}>
        <Typography variant="h6">Create New User</Typography>
        <Box display="flex" gap={2} mt={1}>
          <TextField
            label="Name"
            variant="outlined"
            value={newUserName}
            onChange={e => setNewUserName(e.target.value)}
            fullWidth
          />
          <Button
            variant="contained"
            onClick={handleCreateUser}
            disabled={creating || !newUserName}
          >
            {creating ? <CircularProgress size={24} /> : 'Create'}
          </Button>
        </Box>
      </Box>
      <Box mb={3}>
        <Typography variant="h6">All Users</Typography>
        <List>
          {users.map(u => (
            <React.Fragment key={u.id}>
              <ListItem>
                <ListItemText
                  primary={`${u.name} (ID: ${u.id})`}
                  secondary={accounts[u.id] ? `Balance: ${accounts[u.id].balance}` : 'No account'}
                />
              </ListItem>
              <Divider />
            </React.Fragment>
          ))}
        </List>
      </Box>
      {user && account && (
        <Card>
          <CardContent>
            <Typography variant="h6">Name: {user.name}</Typography>
            <Typography variant="body1">User ID: {user.id}</Typography>
            <Typography variant="body1">Account ID: {account.id}</Typography>
            <Typography variant="body1">Balance: {account.balance}</Typography>
            <Typography variant="body2" color="text.secondary">
              Created: {new Date(account.created_at).toLocaleString()}
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Updated: {new Date(account.updated_at).toLocaleString()}
            </Typography>
          </CardContent>
        </Card>
      )}

      {/* --- Orders Block --- */}
      <Box mt={6}>
        <Typography variant="h5" gutterBottom>Поиск и просмотр заказов</Typography>
        <Box display="flex" gap={2} mb={2}>
          <Button variant="outlined" onClick={fetchAllOrders} disabled={ordersLoading}>Все заказы</Button>
          <TextField
            label="user_id для поиска"
            value={orderSearchUserId}
            onChange={e => setOrderSearchUserId(e.target.value)}
            size="small"
          />
          <Button variant="outlined" onClick={fetchOrdersByUserId} disabled={ordersLoading || !orderSearchUserId}>Заказы по user_id</Button>
          <TextField
            label="order_id для поиска"
            value={orderSearchOrderId}
            onChange={e => setOrderSearchOrderId(e.target.value)}
            size="small"
          />
          <Button variant="outlined" onClick={fetchOrderById} disabled={ordersLoading || !orderSearchOrderId}>Заказ по id</Button>
        </Box>
        {ordersError && <Alert severity="error" sx={{ mb: 2 }}>{ordersError}</Alert>}
        <TableContainer component={Paper} sx={{ mt: 2 }}>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>User ID</TableCell>
                <TableCell>Amount</TableCell>
                <TableCell>Description</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Created</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {orders.map(order => (
                <TableRow key={order.id}>
                  <TableCell>{order.id}</TableCell>
                  <TableCell>{order.user_id}</TableCell>
                  <TableCell>{order.total_amount}</TableCell>
                  <TableCell>{order.description}</TableCell>
                  <TableCell>{order.status}</TableCell>
                  <TableCell>{order.created_at ? new Date(order.created_at).toLocaleString() : ''}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Box>
    </Container>
  );
}

export default Account; 
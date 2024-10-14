<script>
  let todos = [];
  let newTodo = { title: '' };
  let editingTodo = null;
  let isAuthenticated = false;
  let token = localStorage.getItem('token');

  let email = '';
  let password = '';

  let isLoginMode = true;

  const fetchTodos = async () => {
      const response = await fetch('http://localhost:8080/todos', {
          headers: {
              'Authorization': `Bearer ${token}`
          }
      });
      todos = await response.json();
  };

  if (token) {
      isAuthenticated = true;
      fetchTodos();
  } else {
      isAuthenticated = false;
  }

  const login = async (email, password) => {
      const response = await fetch('http://localhost:8080/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
      });
      const data = await response.json();

      if (response.ok) {
          localStorage.setItem('token', data.token);
          token = data.token;
          isAuthenticated = true;
          fetchTodos();
      } else {
          alert('Неверные данные авторизации');
      }
  };

  const register = async (email, password) => {
      const response = await fetch('http://localhost:8080/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
      });
      const data = await response.json();

      if (response.ok) {
          localStorage.setItem('token', data.token);
          token = data.token;
          isAuthenticated = true;
          fetchTodos();
      } else {
          alert('Ошибка регистрации');
      }
  };

  const logout = () => {
      localStorage.removeItem('token');
      isAuthenticated = false;
      todos = [];
  };

  const toggleAuthMode = () => {
      isLoginMode = !isLoginMode;
  };

  // Функции CRUD для задач
  const addTodo = async () => {
      if (newTodo.title.trim() === '') return;

      const response = await fetch('http://localhost:8080/todos', {
          method: 'POST',
          headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`
          },
          body: JSON.stringify(newTodo)
      });

      const addedTodo = await response.json();
      todos = [...todos, addedTodo];
      newTodo = { title: '' };
  };

  const deleteTodo = async (id) => {
      await fetch(`http://localhost:8080/todos/${id}`, {
          method: 'DELETE',
          headers: {
              'Authorization': `Bearer ${token}`
          }
      });

      todos = todos.filter(todo => todo.id !== id);
  };

  const startEdit = (todo) => {
      editingTodo = { ...todo };
  };

  const saveEdit = async () => {
      await fetch(`http://localhost:8080/todos/${editingTodo.id}`, {
          method: 'PUT',
          headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${token}`
          },
          body: JSON.stringify(editingTodo)
      });

      todos = todos.map(todo => (todo.id === editingTodo.id ? editingTodo : todo));
      editingTodo = null;
  };

  const cancelEdit = () => {
      editingTodo = null;
  };
</script>

<main>
  {#if !isAuthenticated}
      {#if isLoginMode}
          <h1>Авторизация</h1>
          <input type="email" bind:value={email} placeholder="Email" />
          <input type="password" bind:value={password} placeholder="Пароль" />
          <button on:click={() => login(email, password)}>Войти</button>
          <p>Нет аккаунта? <a href="#" on:click={toggleAuthMode}>Зарегистрируйтесь</a></p>
      {:else}
          <h1>Регистрация</h1>
          <input type="email" bind:value={email} placeholder="Email" />
          <input type="password" bind:value={password} placeholder="Пароль" />
          <button on:click={() => register(email, password)}>Зарегистрироваться</button>
          <p>Уже есть аккаунт? <a href="#" on:click={toggleAuthMode}>Войдите</a></p>
      {/if}
  {:else}
      <h1>Список дел</h1>
      <input type="text" bind:value={newTodo.title} placeholder="Новая задача">
      <button on:click={addTodo}>Добавить</button>

      <ul>
          {#each todos as todo}
              <li>
                  {#if editingTodo && editingTodo.id === todo.id}
                      <input type="text" bind:value={editingTodo.title} />
                      <button on:click={saveEdit}>Сохранить</button>
                      <button on:click={cancelEdit}>Отмена</button>
                  {:else}
                      {todo.title}
                      <button on:click={() => startEdit(todo)}>Редактировать</button>
                      <button on:click={() => deleteTodo(todo.id)}>Удалить</button>
                  {/if}
              </li>
          {/each}
      </ul>
      <button on:click={logout}>Выйти</button>
  {/if}
</main>

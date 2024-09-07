import "./App.css";
// import { Box } from "@mantine/core";
import useSWR from "swr";
import AddTodo from "./components/AddTodo";

export const ENDPOINT = "http://localhost:4000";

const fetcher = (url) => fetch(`${ENDPOINT}/${url}`).then((r) => r.json());

function App() {
	const { data, mutate } = useSWR("api/todos", fetcher);

	const markTodoDone = async (id) => {
		const updatedTodos = data.map((todo) => (todo.id === id ? { ...todo, done: !todo.done } : todo));

		mutate(updatedTodos, false);

		await fetch(`${ENDPOINT}/api/todos/${id}/done`, {
			method: "PATCH",
		}).then((r) => r.json());

		mutate();
	};

	const deleteTodo = async (id) => {
		await fetch(`${ENDPOINT}/api/todos/${id}`, {
			method: "DELETE",
		});

		const updatedTodos = data.filter((todo) => todo.id !== id);

		mutate(updatedTodos, false);
	};

	return (
		<>
			<div className="p-4 bg-gray-100 rounded-md shadow-md">
				{/* Display data or loading message */}
				{data ? (
					<ul className="space-y-2">
						{data.map((todo) => (
							<li
								key={`todo_list__${todo.id}`}
								className="p-4 bg-white rounded-md shadow-sm hover:bg-gray-50 flex justify-between items-center"
							>
								<div className="flex-1 mx-5">
									<span className="block font-bold">{todo.title}</span>
									<span className="block text-gray-700 mt-1">{todo.body}</span>
								</div>

								<div className="mr-3">
									<button
										onClick={() => markTodoDone(todo.id)}
										className={`${
											todo.done ? "bg-green-500" : "bg-red-500"
										} text-white rounded-full p-1 hover:${
											todo.done ? "bg-green-500" : "bg-red-500"
										} focus:outline-none`}
									>
										{todo.done ? "✓" : "✗"}
									</button>
								</div>

								{/* Delete button */}
								<button
									onClick={() => deleteTodo(todo.id)}
									className="bg-red-500 text-white rounded-full p-1 hover:bg-red-500 focus:outline-none"
								>
									🗑️
								</button>
							</li>
						))}
					</ul>
				) : (
					<p>Loading...</p>
				)}
			</div>

			{/* AddTodo component */}
			<AddTodo mutate={mutate} data={data} />
		</>
	);
}

export default App;

import "./App.css";
// import { Box } from "@mantine/core";
import useSWR from "swr";
import AddTodo from "./components/AddTodo";

export const ENDPOINT = "http://localhost:4000";

const fetcher = (url) => fetch(`${ENDPOINT}/${url}`).then((r) => r.json());

function App() {
	const { data, mutate } = useSWR("api/todos", fetcher);

	const markTodoDone = async (id) => {
		const updated = await fetch(`${ENDPOINT}/api/todos/${id}/done`, {
			method: "PATCH",
		}).then((r) => r.json());

		mutate(updated);
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
								className="p-2 bg-white rounded-md shadow-sm hover:bg-gray-50 flex justify-between items-center"
							>
								<span>{todo.title}</span>
								<button
									onClick={() => markTodoDone(todo.id)}
									className={`${todo.done ? "bg-green-500" : "bg-red-500"} text-white rounded-full p-2 hover:${
										todo.done ? "bg-green-600" : "bg-red-600"
									} focus:outline-none`}
								>
									{todo.done ? "✓" : "✗"}
								</button>
							</li>
						))}
					</ul>
				) : (
					<p>Loading...</p>
				)}
			</div>

			{/* AddTodo component */}
			<AddTodo mutate={mutate} />
		</>
	);
}

export default App;

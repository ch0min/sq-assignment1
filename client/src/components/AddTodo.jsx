import { useState } from "react";
import { ENDPOINT } from "../App";
// import { useForm } from "@mantine/hooks";
// import { Modal, Group, Button } from "@mantine/core";

function AddTodo({ mutate, data }) {
	const [open, setOpen] = useState(false);
	const [formValues, setFormValues] = useState({
		title: "",
		body: "",
		category: "",
		// deadline: "",
	});

	const handleChange = (e) => {
		setFormValues({
			...formValues,
			[e.target.name]: e.target.value,
		});
	};

	const createTodo = async () => {
		const newTodo = await fetch(`${ENDPOINT}/api/todos`, {
			method: "POST",
			headers: {
				"Content-Type": "application/json",
			},
			body: JSON.stringify(formValues),
		}).then((r) => r.json());

		mutate([...data, newTodo], false);

		setFormValues({
			title: "",
			body: "",
			category: "",
			// deadline: "",
		});

		setOpen(false);
	};

	return (
		<>
			{/* Modal */}
			{open && (
				<div className="fixed inset-0 flex items-center justify-center z-50 bg-black bg-opacity-50">
					<div className="bg-white rounded-lg p-6 w-1/3">
						<h2 className="text-lg font-semibold mb-4">Create Todo</h2>
						<form onSubmit={createTodo}>
							<div className="mb-4">
								<label htmlFor="title" className="block text-sm font-medium text-gray-700">
									Title
								</label>
								<input
									id="title"
									name="title"
									type="text"
									value={formValues.title}
									onChange={handleChange}
									placeholder="What is your todo?"
									required
									className="mt-1 p-2 block w-full border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
								/>
							</div>
							<div className="mb-4">
								<label htmlFor="category" className="block text-sm font-medium text-gray-700">
									Category
								</label>
								<input
									id="category"
									name="category"
									type="text"
									value={formValues.category}
									onChange={handleChange}
									placeholder="Enter category"
									className="mt-1 p-2 block w-full border border-gray-300 rounded-md"
								/>
							</div>
							<div className="mb-4">
								<label htmlFor="body" className="block text-sm font-medium text-gray-700">
									Body
								</label>
								<textarea
									id="body"
									name="body"
									rows="4"
									value={formValues.body}
									onChange={handleChange}
									placeholder="Tell me more..."
									required
									className="mt-1 p-2 block w-full border border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
								/>
							</div>

							<div className="flex justify-end">
								<button
									type="button"
									onClick={() => setOpen(false)}
									className="mr-2 bg-gray-300 hover:bg-gray-400 text-gray-700 px-4 py-2 rounded-md"
								>
									Cancel
								</button>
								<button type="submit" className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-md">
									Save
								</button>
							</div>
						</form>
					</div>
				</div>
			)}

			{/* Button to open modal */}
			<div className="flex justify-center my-4">
				<button className="bg-blue-500 text-white font-bold py-2 px-4 rounded-md" onClick={() => setOpen(true)}>
					ADD TODO
				</button>
			</div>
		</>
	);
}

export default AddTodo;

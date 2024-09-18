import React, { useState } from "react";
import { ENDPOINT } from "../App";

function EditTodo({ todo, onSave, onCancel }) {
	const [formValues, setFormValues] = useState({
		title: todo.title,
		body: todo.body,
		category: todo.category,
		deadline: todo.deadline ? new Date(todo.deadline).toISOString().substring(0, 10) : "",
	});

	const handleChange = (e) => {
		setFormValues({
			...formValues,
			[e.target.name]: e.target.value,
		});
	};

	const handleSubmit = (e) => {
		e.preventDefault();
		onSave({ ...todo, ...formValues });
	};

	return (
		<div className="p-4 bg-white rounded-md shadow-sm">
			<form onSubmit={handleSubmit}>
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
						required
						className="mt-1 p-2 block w-full border border-gray-300 rounded-md shadow-sm"
					/>
				</div>
				<div className="mb-4">
					<label htmlFor="body" className="block text-sm font-medium text-gray-700">
						Body
					</label>
					<textarea
						id="body"
						name="body"
						value={formValues.body}
						onChange={handleChange}
						rows="3"
						className="mt-1 p-2 block w-full border border-gray-300 rounded-md shadow-sm"
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
						className="mt-1 p-2 block w-full border border-gray-300 rounded-md"
					/>
				</div>

				<div className="flex justify-end">
					<button
						type="button"
						onClick={onCancel}
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
	);
}

export default EditTodo;

// App.test.js
import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";
import App from "../App";
import AddTodo from "../components/AddTodo";

// Mock the SWR hook to simulate the fetching of todos
vi.mock("swr", () => ({
	default: () => ({
		data: [
			{ id: 1, title: "Test Todo 1", body: "Test Body 1", category: "Work", done: false },
			{ id: 2, title: "Test Todo 2", body: "Test Body 2", category: "Personal", done: true },
		],
		mutate: vi.fn(),
	}),
}));

// Mock the global fetch function to prevent real API calls
global.fetch = vi.fn(() =>
	Promise.resolve({
		json: () => Promise.resolve({}),
	})
);

// Tests
test("displays a list of todos correctly", () => {
	render(<App />);

	// Check if the todos are displayed on the screen
	expect(screen.getByText("Test Todo 1")).toBeInTheDocument();
	expect(screen.getByText("Test Body 1")).toBeInTheDocument();
	expect(screen.getByText("Test Todo 2")).toBeInTheDocument();
	expect(screen.getByText("Test Body 2")).toBeInTheDocument();

	// Check if the "done" and "not done" indicators are present
	expect(screen.getByText("☒")).toBeInTheDocument(); // Not done
	expect(screen.getByText("☑")).toBeInTheDocument(); // Done
});

test("updates form input values correctly when user types", () => {
	render(<AddTodo />);

	const addButton = screen.getByText("ADD TODO");
	fireEvent.click(addButton);

	const titleInput = screen.getByPlaceholderText("What is your todo?");
	const bodyInput = screen.getByPlaceholderText("Tell me more...");
	const categoryInput = screen.getByPlaceholderText("Enter category");

	// Simulate user typing into the input fields
	fireEvent.change(titleInput, { target: { value: "New Todo Title" } });
	fireEvent.change(bodyInput, { target: { value: "New Todo Description" } });
	fireEvent.change(categoryInput, { target: { value: "New Todo Category" } });

	// Assert that the input values are updated correctly
	expect(titleInput.value).toBe("New Todo Title");
	expect(bodyInput.value).toBe("New Todo Description");
	expect(categoryInput.value).toBe("New Todo Category");
});

test("removes a todo when delete button is clicked", async () => {
	// Render the App component
	render(<App />);

	// Ensure that both todos are initially rendered on the screen
	expect(screen.getByText("Test Todo 1")).toBeInTheDocument();
	expect(screen.getByText("Test Todo 2")).toBeInTheDocument();

	// Simulate clicking the delete button for the first todo
	const deleteButton = screen.getAllByText("✘")[0];
	fireEvent.click(deleteButton);

	// Verify that the fetch function was called with the correct URL and method (DELETE) for the first todo
	expect(global.fetch).toHaveBeenCalledWith(
		expect.stringContaining("/api/todos/1"),
		expect.objectContaining({ method: "DELETE" })
	);

	// Check that the mocked fetch function was called (indicating the deletion API call occurred)
	expect(vi.mocked(fetch)).toHaveBeenCalled();
});

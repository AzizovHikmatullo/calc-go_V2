document.addEventListener("DOMContentLoaded", () => {
    const API_BASE_URL = "http://localhost:8080";

    const expressionForm = document.getElementById("expressionForm");
    expressionForm.addEventListener("submit", async (e) => {
        e.preventDefault();
        const expression = document.getElementById("expressionInput").value;

        try {
            const response = await fetch(`${API_BASE_URL}/api/v1/calculate`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ expression }),
            });

            if (!response.ok) {
                throw new Error(`Error: ${response.status}`);
            }

            const data = await response.json();
            alert(`Выражение успешно отрпавлено! ID: ${data.id}`);
        } catch (error) {
            alert(`Не удалось отправить выражение: ${error.message}`);
        }
    });

    const refreshListButton = document.getElementById("refreshListButton");
    const expressionsList = document.getElementById("expressionsList");
    refreshListButton.addEventListener("click", async () => {
        try {
            const response = await fetch(`${API_BASE_URL}/api/v1/expressions`);
            if (!response.ok) {
                throw new Error(`Ошибка: ${response.status}`);
            }

            const data = await response.json();
            expressionsList.innerHTML = "";

            data.expressions.forEach((expr) => {
                const listItem = document.createElement("li");
                listItem.textContent = `ID: ${expr.id}, Выражение: ${expr.expression}, Статус: ${expr.status}, Результат: ${expr.result || "N/A"}`;
                expressionsList.appendChild(listItem);
            });
        } catch (error) {
            alert(`Не удалось найти выражения: ${error.message}`);
        }
    });

    const getDetailsButton = document.getElementById("getDetailsButton");
    const expressionIdInput = document.getElementById("expressionIdInput");
    const expressionDetails = document.getElementById("expressionDetails");
    getDetailsButton.addEventListener("click", async () => {
        const expressionId = expressionIdInput.value.trim();

        if (!expressionId) {
            alert("Пожалуйста, введите ID выражения");
            return;
        }

        try {
            const response = await fetch(`${API_BASE_URL}/api/v1/expressions/${expressionId}`);
            if (!response.ok) {
                throw new Error(`Ошибка: ${response.status}`);
            }

            const data = await response.json();
            expressionDetails.textContent = JSON.stringify(data.expression, null, 2);
        } catch (error) {
            alert(`Не удалось найти выражение: ${error.message}`);
        }
    });
});
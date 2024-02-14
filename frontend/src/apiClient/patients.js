export const fetchPatientImages = async () => {
  const url = `http://localhost:9797/v1/patient/images/list`;

  const payload = {
    patient: {
      id: "1",
    },
    queryOption: {
      filters: {
        tags: {
          tag: ["tag1"],
        },
        filterFields: {
          comparisonOperator: "COMPARISON_UNKNOWN",
        },
        logicalOperator: "AND",
        sortOperator: "DEFAULT",
      },
      pagination: {
        pageToken: "string",
        pageSize: 0,
      },
    },
  };

  const options = {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(payload),
  };

  try {
    const response = await fetch(url, options);
    if (!response.ok) {
      throw new Error("Network response was not ok");
    }
    const serializedResponse = await response.json();
    return serializedResponse;
  } catch (error) {
    console.error("Error:", error);
  }
};


export const uploadFile = async (selectedFile, fileName, contentType) => {
	const url = `http://localhost:9797/v1/patient/images/upload`;

	const formData = new FormData();
	formData.append("file", selectedFile);

	console.log("formData", formData.get(fileName));

	const payload = 	{
		patientImage: {
			patientId: "1",
			image: {
				name: string,
				description: "string",
				tags: [
					"tag1"
				]
			}
		},
		tags: {
			tag: []
		},
		filePath: ""
	}

	formData.append("payload", payload)

	try {
		const response = await fetch(url, {
			method: "POST",
			headers: {
				// "Content-Type": "multipart/form-data",
			},
			body: formData,
		});
		if (!response.ok) {
			throw new Error("Network response was not ok");
		}
		const data = await response.json();
		console.log("Success:", data);
	} catch (error) {
		console.error("Error:", error);
	}
};

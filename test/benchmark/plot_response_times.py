import matplotlib.pyplot as plt
import pandas as pd
import sys

if len(sys.argv) != 3:
    print("FOLLOW -> py plot_response_times.py <input_csv> <output_image>")
    sys.exit(1)

input_csv = sys.argv[1]
output_image = sys.argv[2]

# Load the data.
data = pd.read_csv(input_csv)

# Plot the graph.
plt.figure(figsize=(10, 6))
plt.plot(data["RequestNumber"], data["ResponseTime(ms)"], label="Response Time")
plt.xlabel("Request #")
plt.ylabel("Response Time (ms)")
plt.title("Response Time Distribution")
plt.legend()
plt.grid()
plt.savefig(output_image)

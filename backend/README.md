# hdviz

Using caution:
Make sure you are not runnning any other operations when performing the build, as the build process would read all the files in your system and generate a foldermap.json for vizualization.


Once you run sudo go run main.go, there will be a foldermap.json generated in your frontend directory.
foldermap.json would contain the series of dictionary inputs:
```json
{
 "/Users/Toyakki": {
  "FolderName": "Toyakki",
  "Size": 274621020071,
  "TopKChildren": [
   "/Users/Toyakki/.ghcup",
   "/Users/Toyakki/LinuxVM",
   "/Users/Toyakki/.local",
   "/Users/Toyakki/Pictures",
   "/Users/Toyakki/.npm",
   "/Users/Toyakki/.ghcup.bak.20250916174528",
   "/Users/Toyakki/.ollama",
   "/Users/Toyakki/miniconda3",
   "/Users/Toyakki/.cache",
   "/Users/Toyakki/Library"
  ]
 }, ...
}
```

TODOs:
1. Offline build pathways
- Convert folderMap to JSON format and store it in a frontend

- Add an autodelete function to remove this JSON once the frontend has loaded it to save user's storage space



2. Online build 
- Add some condition check to determine whether the user can perform online or offline build, and then trigger the corresponding build pathway. Default is offline, but if the user does not have an enough space to generate the folderMap JSON, then trigger the online build pathway.
- Expose a Restful endpoint to serve the folderMap JSON to the frontend

Future:
- Create a duplicate of this repo and store it to the public github, make sure to remove foldermap.json. 
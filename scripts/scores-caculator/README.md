# Scores Module Documentation

## How to Use the Scores Module

### Build and Installation

First, you need to build the project using the `make` command. Run the following command from the project's root directory to compile and install:

```
make
```

### Execution Command

After building, you can run the Scores module using the following command:

```
./bin/gen_scores -config=config.json
```

### Parameter Explanation

- `-config`: Specifies the path to the configuration file. The configuration file typically includes database connection details like host, port, username, password, etc. The default is `config.json`, but you can provide a different file if needed.

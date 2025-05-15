# Concepts

**yams** has a number of proprietary concepts which help to model data and simulations.
Understanding these concepts is not required to use the tooling, but can help make sense of the
documentation and errors associated with the tool.

### Sources

Getting data into **yams** revolves around the concept of a ***Source***, which contains definitions
for principals, resources, policies, accounts, etc. **Sources** are read periodically by the
**yams** server to keep data fresh.

### Entities

An **Entity** is a loose term to describe a _thing_ associated with simulation. Included in the
definition of **Entities** are:

- Principals
- Groups
- Resources
- Policies
- Tags
- Accounts
- Org Nodes

### Universe

A **Universe** is a container which holds all known **Entities**. The most commonly used
**Universe** would be the one representing baseline reality as defined by a set of **Sources**,
however full or partial alternative **Universes** can be constructed as desired.

### Overlay

An **Overlay** or **Overlay Universe** is similarly a container of **Entities**, but has the purpose
of redefining or overriding **Entity** definitions defined in a base **Universe**. Priority is
always given to the **Overlay** when resolving configurations.

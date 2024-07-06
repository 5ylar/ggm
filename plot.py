import plotly.express as px
import pandas as pd

df = pd.read_csv("points.csv")

# df = pd.DataFrame.from_records(
#     [
#         (13.971514, 100.403076),
#         (13.967022423579404, 100.403076),
#     ],
#     columns=['Lat', 'Long'],
# )

print(df)


df.dropna(
    axis=0,
    how='any',
    # thresh=None,
    subset=None,
    inplace=True
)

color_scale = [(0, 'orange'), (1, 'red')]

fig = px.scatter_mapbox(df,
                        lat="Lat",
                        lon="Long",
                        # hover_name="Address",
                        # hover_data=["Address", "Listed"],
                        # color="Listed",
                        color_continuous_scale=color_scale,
                        # size="Listed",
                        zoom=8,
                        height=800,
                        width=800)

fig.update_layout(mapbox_style="open-street-map")
fig.update_layout(margin={"r": 0, "t": 0, "l": 0, "b": 0})
fig.show()
